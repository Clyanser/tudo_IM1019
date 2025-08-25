package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync"
	"time"
	"tudo_IM1019/common/models/ctype"
	"tudo_IM1019/tudoIM_chat/chat_models"
	"tudo_IM1019/tudoIM_file/file_rpc/types/file_rpc"
	"tudo_IM1019/tudoIM_user/user_models"
	"tudo_IM1019/tudoIM_user/user_rpc/types/user_rpc"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/svc"
	"tudo_IM1019/tudoIM_chat/chat_api/internal/types"

	"tudo_IM1019/common/response"
)

type UserWsInfo struct {
	UserInfo user_models.UserModel //用户信息
	Conn     *websocket.Conn       //用户的ws 连接对象
}

var UseOnlineWsMap = map[uint]UserWsInfo{}
var userWsMapMutex sync.RWMutex // 读写锁

func chatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := httpx.ParseHeaders(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		logx.Infof("【连接开始】用户 %d 正在建立 WebSocket 连接，当前 UserWsMap 状态: %v", req.UserID, UseOnlineWsMap)
		//ws 升级
		var upGrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				//鉴权 true 表示旅行 false 表示拦截
				return true
			},
		}
		conn, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		defer func() {
			_ = conn.Close() //连接对象断开
			//
			userWsMapMutex.Lock()
			delete(UseOnlineWsMap, req.UserID)
			userWsMapMutex.Unlock()
			svcCtx.RDB.HDel("online", fmt.Sprintf("%d", req.UserID))
		}()
		//调用户服务，获取当前用户信息
		res, err := svcCtx.UserRpc.UserInfo(context.Background(), &user_rpc.UserInfoRequest{
			UserId: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		var userInfo user_models.UserModel
		err = json.Unmarshal(res.Data, &userInfo)
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		var userWsInfo = UserWsInfo{
			UserInfo: userInfo,
			Conn:     conn,
		}

		// 如果该用户已在线，先关闭旧连接
		oldConn, exists := UseOnlineWsMap[req.UserID]
		if exists {
			logx.Infof("用户 %d 重复连接，关闭旧连接", req.UserID)
			_ = oldConn.Conn.Close()
			//加锁删除
			userWsMapMutex.Lock()
			delete(UseOnlineWsMap, req.UserID)
			userWsMapMutex.Unlock()
		}
		//设置新连接
		userWsMapMutex.Lock()
		UseOnlineWsMap[req.UserID] = userWsInfo
		userWsMapMutex.Unlock()
		logx.Infof("【连接建立】用户 %d 已存入 UserWsMap", req.UserID)
		//把在线的用户存进公共的地方（redis）
		svcCtx.RDB.HSet("online", fmt.Sprintf("%d", req.UserID), req.UserID)
		//遍历当前在线的好友，在线的就给他发信息（我已上线）

		//先把所有在线的用户id 取出来，以及待确认的用户的id ,然后传到用户rpc 服务中

		//在rpc 服务中，去判断哪些用户是好友关系

		//查自己的好友列表，返回用户id列表
		//if userInfo.UserConfModel.FriendOnline {
		//如果用户开启了好友上线提醒
		//查一下自己的好友是否上线了
		list, err := svcCtx.UserRpc.FriendList(context.Background(), &user_rpc.FriendListRequest{
			User: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}
		//读操作上锁
		userWsMapMutex.RLock()
		for _, v := range list.FriendList {
			friend, ok := UseOnlineWsMap[uint(v.UserId)]
			if ok {
				text := fmt.Sprintf("好友 %s 已经上线", UseOnlineWsMap[req.UserID].UserInfo.Nickname)
				//判断好友是否开启了好友上线提醒
				if friend.UserInfo.UserConfModel.FriendOnline {
					//好友上线了
					err := friend.Conn.WriteMessage(websocket.TextMessage, []byte(text))
					if err != nil {
						logx.Error(err)
						response.Response(r, w, nil, err)
						return
					}
				}
			}
		}
		userWsMapMutex.RUnlock()
		//}
		for {
			//消息类型，消息，错误
			_, p, err1 := conn.ReadMessage()
			if err1 != nil {
				//用户断开聊天
				fmt.Println(err)
				break
			}
			var request chatRequest
			err2 := json.Unmarshal(p, &request)
			if err2 != nil {
				//用户乱发消息
				logx.Error(err2)
				SendTipErrMsg(conn, "参数解析失败")
				continue
			}
			if request.RevUserID != req.UserID { //实现自己能跟自己聊天
				// 判断是否是你的好友
				isFriendRes, err := svcCtx.UserRpc.IsFriend(context.Background(), &user_rpc.IsFriendRequest{
					User1: uint32(req.UserID),
					User2: uint32(request.RevUserID),
				})
				if err != nil {
					logx.Error(err)
					SendTipErrMsg(conn, "网络错误")
					continue
				}
				if !isFriendRes.IsFriend {
					SendTipErrMsg(conn, "你们还不是好友")
					continue
				}
			}
			//判断消息类型
			switch request.Msg.Type {
			case ctype.TextMsgType:
				if request.Msg.TextMsg == nil {
					SendTipErrMsg(conn, "请输入消息内容")
					continue
				}
				if request.Msg.TextMsg.Content == "" {
					SendTipErrMsg(conn, "请输入消息内容")
					continue
				}
				//判断是否是文件类型
			case ctype.FileMsgType:
				if request.Msg.FileMsg == nil {
					logx.Error("文件消息体为空")
					return
				}
				logx.Infof("原始 File URL: '%s'", request.Msg.FileMsg.Url)
				nameList := strings.Split(request.Msg.FileMsg.Url, "/")
				var uuid string
				if len(nameList) == 0 {
					SendTipErrMsg(conn, "请上传文件")
					continue
				}
				//fmt.Println(nameList)
				uuid = nameList[len(nameList)-1]
				//fmt.Println(uuid) *调试信息*
				//	是文件类型，请求文件rpc服务 *调试信息*
				fileResponse, err3 := svcCtx.FileRpc.FileInfo(context.Background(), &file_rpc.FileInfoRequest{
					FileId: uuid,
				})
				if err3 != nil {
					logx.Error(err3)
					SendTipErrMsg(conn, err3.Error())
					continue
				}
				request.Msg.FileMsg.Title = fileResponse.FileName
				request.Msg.FileMsg.Type = fileResponse.FileType
				request.Msg.FileMsg.Size = fileResponse.FileSize
			case ctype.RecallMsgType:
				if request.Msg.RecallMsg == nil {
					logx.Error("撤回消息体为空")
					continue
				}
				//	撤回消息ID必填
				if request.Msg.RecallMsg.MsgID == 0 {
					logx.Info("请填写撤回的消息ID")
					SendTipErrMsg(conn, "撤回失败")
					continue
				}
				//  判断有没有这条消息
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, "id = ?", request.Msg.RecallMsg.MsgID).Error
				if err != nil {
					logx.Error(err)
					SendTipErrMsg(conn, "撤回失败")
					continue
				}
				//已经是撤回消息的，不能再撤回了
				if msgModel.MsgType == ctype.RecallMsgType {
					SendTipErrMsg(conn, "已经撤回消息了，您无法再次撤回这条消息")
					continue
				}

				//  判断是不是自己发的
				if msgModel.SendUserID != req.UserID {
					//	不是自己发的
					SendTipErrMsg(conn, "只能撤回自己的消息")
					continue
				}
				//	撤回逻辑
				//  判断撤回的时间小于发送时间两分钟
				nowTime := time.Now()
				validTime := nowTime.Sub(msgModel.CreatedAt)
				if validTime >= time.Minute*2 {
					SendTipErrMsg(conn, "撤回失败,已过撤回时间")
					continue
				}
				//收到撤回请求后，服务端把原消息修改为撤回提示消息，并记录原消息
				//通知前端收发双方刷新聊天记录
				content := fmt.Sprintf("%s 撤回了一条消息", userInfo.Nickname)
				if userInfo.UserConfModel.RecallMessage != nil {
					logx.Info("使用用户自定义的撤回回复")
					content = content + *userInfo.UserConfModel.RecallMessage
				}
				//自己备份，解决循环引用
				originMsg := msgModel.Msg
				originMsg.RecallMsg = nil //这里可能会出现循环引用，拷贝原消息，并把撤回消息置空

				err = svcCtx.DB.Model(&msgModel).Updates(chat_models.ChatModel{
					MsgPreview: "[撤回消息 -]" + content,
					MsgType:    ctype.RecallMsgType,
					Msg: ctype.Msg{
						Type: ctype.RecallMsgType,
						RecallMsg: &ctype.RecallMsg{
							Notice:    content,
							MsgID:     request.Msg.RecallMsg.MsgID,
							OriginMsg: &originMsg,
						},
					},
				}).Error
				if err != nil {
					logx.Error(err)
					continue
				}
				//构造响应
				request.Msg.RecallMsg.Notice = content
				//request.Msg.RecallMsg.OriginMsg = &originMsg
				//	把原消息置为空
			case ctype.ReplyMsgType:
				// 回复消息
				// 先校验
				if request.Msg.ReplyMsg == nil || request.Msg.ReplyMsg.MsgId == 0 {
					SendTipErrMsg(conn, "回复消息id必填")
					continue
				}

				// 找这个原消息
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, request.Msg.ReplyMsg.MsgId).Error
				if err != nil {
					SendTipErrMsg(conn, "消息不存在")
					continue
				}

				// 不能回复撤回消息
				if msgModel.MsgType == ctype.RecallMsgType {
					SendTipErrMsg(conn, "该消息已撤回")
					continue
				}

				// 回复的这个消息，必须是你自己或者当前和你聊天这个人发出来的

				// 原消息必须是 当前你要和对方聊的  原消息就会有一个 发送人id和接收人id，  我们聊天也会有一个发送人id和接收人id
				// 因为回复消息可以回复自己的，也可以回复别人的
				// 如果回复只能回复别人的？那么条件怎么写?
				fmt.Println(msgModel.SendUserID, msgModel.RevUserID)
				fmt.Println(req.UserID, request.RevUserID)
				if !((msgModel.SendUserID == req.UserID && msgModel.RevUserID == request.RevUserID) ||
					(msgModel.SendUserID == request.RevUserID && msgModel.RevUserID == req.UserID)) {
					SendTipErrMsg(conn, "只能回复自己或者对方的消息")
					continue
				}

				userBaseInfo, err := svcCtx.UserRpc.UserBaseInfo(context.Background(), &user_rpc.UserBaseInfoRequest{
					UserId: uint32(msgModel.SendUserID),
				})
				if err != nil {
					logx.Error(err)
					return
				}

				request.Msg.ReplyMsg.MsgContent = &msgModel.Msg
				request.Msg.ReplyMsg.UserId = msgModel.SendUserID
				request.Msg.ReplyMsg.UserNickName = userBaseInfo.NickName
				request.Msg.ReplyMsg.OriginMsgDate = msgModel.CreatedAt

			case ctype.QuoteMsgType:
				//	引用消息
				if request.Msg.QuoteMsg == nil {
					logx.Error("引用消息体为空")
					continue
				}
				if request.Msg.QuoteMsg.MsgId == 0 {
					logx.Error("引用的消息id必填")
					continue
				}
				//获取原消息
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, request.Msg.QuoteMsg.MsgId).Error
				if err != nil {
					logx.Error(err)
					SendTipErrMsg(conn, "消息发送失败")
					continue
				}
				//获取用户基本信息
				userBaseInfo, err := svcCtx.UserRpc.UserBaseInfo(context.Background(), &user_rpc.UserBaseInfoRequest{
					UserId: uint32(msgModel.SendUserID),
				})
				if err != nil {
					logx.Error(err)
					return
				}
				//	构造响应
				request.Msg.QuoteMsg.MsgContent = &msgModel.Msg
				request.Msg.QuoteMsg.UserId = msgModel.SendUserID
				request.Msg.QuoteMsg.UserNickName = userBaseInfo.NickName
				request.Msg.QuoteMsg.OriginMsgDate = msgModel.CreatedAt
			default:
				logx.Error("非法的消息type")
				SendTipErrMsg(conn, "消息发送失败，请重试！")
				continue
			}

			//先入库
			msgID := InsertMsgByChat(svcCtx.DB, request.RevUserID, req.UserID, request.Msg)
			//看看目标用户在不在线  给发送双方都要发消息
			SendMsgByUser(svcCtx, request.RevUserID, req.UserID, request.Msg, msgID)
		}
	}
}

type chatRequest struct {
	RevUserID uint      `json:"rev_user_id"`
	Msg       ctype.Msg `json:"msg"`
}

type chatResponse struct {
	ID       uint64         `json:"id"`
	IsMe     bool           `json:"is_me"`
	RevUser  ctype.UserInfo `json:"revUser"`
	SendUser ctype.UserInfo `json:"sendUser"`
	Msg      ctype.Msg      `json:"msg"`
	CreateAt time.Time      `json:"created_at"`
}

func InsertMsgByChat(db *gorm.DB, revUserID uint, sendUserID uint, msg ctype.Msg) (msgID uint64) {
	switch msg.Type {
	case ctype.RecallMsgType:
		fmt.Println("撤回消息不入库")
		return
	default:

	}
	chatModel := chat_models.ChatModel{
		RevUserID:  revUserID,
		SendUserID: sendUserID,
		MsgType:    msg.Type,
		Msg:        msg,
	}
	chatModel.MsgPreview = chatModel.MsgPreviewMethod()
	err := db.Create(&chatModel).Error
	if err != nil {
		logx.Error(err)
		//上锁
		userWsMapMutex.RLock()
		sendUser, ok := UseOnlineWsMap[sendUserID]
		userWsMapMutex.RUnlock()
		if !ok {
			return
		}
		SendTipErrMsg(sendUser.Conn, "超过发送限制！")
	}
	return chatModel.ID
}

// SendTipErrMsg 发送错误提示消息
func SendTipErrMsg(conn *websocket.Conn, msg string) {
	resp := chatResponse{
		Msg: ctype.Msg{
			Type: ctype.TipMsgType,
			TipMsg: &ctype.TipMsg{
				Status:  "error",
				Content: msg,
			},
		},
		CreateAt: time.Now(),
	}
	bytes, _ := json.Marshal(resp)
	_ = conn.WriteMessage(websocket.TextMessage, bytes)
}

// SendMsgByUser 发给谁 谁发的 消息
func SendMsgByUser(svcCtx *svc.ServiceContext, revUserID uint, sendUserID uint, msg ctype.Msg, msgID uint64) {
	//上锁
	userWsMapMutex.RLock()
	defer userWsMapMutex.RUnlock()

	revUser, ok1 := UseOnlineWsMap[revUserID]
	sendUser, ok2 := UseOnlineWsMap[sendUserID]
	resp := chatResponse{
		ID:       msgID,
		Msg:      msg,
		CreateAt: time.Now(),
	}

	if ok1 && ok2 && sendUserID == revUserID {
		resp.IsMe = true
		//自己给自己发消息
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: revUser.UserInfo.Nickname,
			Avatar:   revUser.UserInfo.Avatar,
		}
		resp.SendUser = ctype.UserInfo{
			ID:       sendUserID,
			NickName: sendUser.UserInfo.Nickname,
			Avatar:   sendUser.UserInfo.Avatar,
		}
		bytDate, _ := json.Marshal(resp)
		err := revUser.Conn.WriteMessage(websocket.TextMessage, bytDate)
		if err != nil {
			logx.Error(err)
			return
		}
		return
	}
	//不管怎么样，都要给发送者回传消息
	//如果接收者不在线，那么我就要去拿接收者的用户信息

	// === 给发送者回传消息 ===
	if ok2 {
		resp.IsMe = true
		resp.SendUser = ctype.UserInfo{
			ID:       sendUserID,
			NickName: sendUser.UserInfo.Nickname,
			Avatar:   sendUser.UserInfo.Avatar,
		}
		sendBytes, _ := json.Marshal(resp)
		_ = sendUser.Conn.WriteMessage(websocket.TextMessage, sendBytes)
	}

	// === 给接收者发送消息（仅当在线）===
	if ok1 {
		resp.IsMe = false
		// 接收者在线，使用其连接信息
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: revUser.UserInfo.Nickname,
			Avatar:   revUser.UserInfo.Avatar,
		}
		revBytes, _ := json.Marshal(resp)
		_ = revUser.Conn.WriteMessage(websocket.TextMessage, revBytes)
	} else {
		// 接收者不在线，需要拿接收者的信息 → 存离线消息（建议实现）
		userBaseInfo, err := svcCtx.UserRpc.UserBaseInfo(context.Background(), &user_rpc.UserBaseInfoRequest{
			UserId: uint32(revUserID),
		})
		if err != nil {
			logx.Error(err)
			return
		}
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: userBaseInfo.NickName,
			Avatar:   userBaseInfo.Avatar,
		}
		//不在线，那就不需要发送了，消息入库即可
		InsertMsgByChat(svcCtx.DB, revUserID, sendUserID, msg)
		logx.Infof("User %d is offline, storing offline message with ID %d", revUserID, msgID)
		// storeOfflineMessage(svcCtx, revUserID, resp)
	}
}
