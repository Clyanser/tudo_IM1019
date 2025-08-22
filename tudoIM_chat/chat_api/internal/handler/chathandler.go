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

var UserWsMap = map[uint]UserWsInfo{}
var userWsMapMutex sync.RWMutex // 读写锁

func chatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := httpx.ParseHeaders(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		logx.Infof("【连接开始】用户 %d 正在建立 WebSocket 连接，当前 UserWsMap 状态: %v", req.UserID, UserWsMap)
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
			delete(UserWsMap, req.UserID)
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
		oldConn, exists := UserWsMap[req.UserID]
		if exists {
			logx.Infof("用户 %d 重复连接，关闭旧连接", req.UserID)
			_ = oldConn.Conn.Close()
			//加锁删除
			userWsMapMutex.Lock()
			delete(UserWsMap, req.UserID)
			userWsMapMutex.Unlock()
		}
		//设置新连接
		userWsMapMutex.Lock()
		UserWsMap[req.UserID] = userWsInfo
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
			friend, ok := UserWsMap[uint(v.UserId)]
			if ok {
				text := fmt.Sprintf("好友 %s 已经上线", UserWsMap[req.UserID].UserInfo.Nickname)
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
			//判断是否是文件类型
			switch request.Msg.Type {
			case ctype.FileMsgType:
				nameList := strings.Split(request.Msg.FileMsg.Url, "/")
				var uuid string
				if len(nameList) == 0 {
					SendTipErrMsg(conn, "请上传文件")
					continue
				}
				fmt.Println(nameList)
				uuid = nameList[len(nameList)-1]
				fmt.Println(uuid)
				//	是文件类型，请求文件rpc服务
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
			default:
			}

			//先入库
			InsertMsgByChat(svcCtx.DB, request.RevUserID, req.UserID, request.Msg)
			//看看目标用户在不在线
			SendMsgByUser(request.RevUserID, req.UserID, request.Msg)
		}
	}
}

type chatRequest struct {
	RevUserID uint      `json:"rev_user_id"`
	Msg       ctype.Msg `json:"msg"`
}

type chatResponse struct {
	RevUser  ctype.UserInfo `json:"revUser"`
	SendUser ctype.UserInfo `json:"sendUser"`
	Msg      ctype.Msg      `json:"msg"`
	CreateAt time.Time      `json:"createAt"`
}

func InsertMsgByChat(db *gorm.DB, revUserID uint, sendUserID uint, msg ctype.Msg) {
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
		sendUser, ok := UserWsMap[sendUserID]
		userWsMapMutex.RUnlock()
		if !ok {
			return
		}
		SendTipErrMsg(sendUser.Conn, "超过发送限制！")
	}
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
func SendMsgByUser(revUserID uint, sendUserID uint, msg ctype.Msg) {
	//上锁
	userWsMapMutex.RLock()
	defer userWsMapMutex.RUnlock()

	revUser, ok := UserWsMap[revUserID]
	if !ok {
		return
	}
	sendUser, ok := UserWsMap[sendUserID]
	if !ok {
		return
	}
	resp := chatResponse{
		RevUser: ctype.UserInfo{
			ID:       revUserID,
			NickName: revUser.UserInfo.Nickname,
			Avatar:   revUser.UserInfo.Avatar,
		},
		SendUser: ctype.UserInfo{
			ID:       sendUserID,
			NickName: sendUser.UserInfo.Nickname,
			Avatar:   sendUser.UserInfo.Avatar,
		},
		Msg:      msg,
		CreateAt: time.Now(),
	}
	bytdate, _ := json.Marshal(resp)
	revUser.Conn.WriteMessage(websocket.TextMessage, bytdate)
}
