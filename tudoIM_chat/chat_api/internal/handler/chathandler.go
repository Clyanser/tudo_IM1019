package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
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

func chatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := httpx.ParseHeaders(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}
		//logx.Infof("【连接开始】用户 %d 正在建立 WebSocket 连接，当前 UserWsMap 状态: %v", req.UserID, UserWsMap)
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
			delete(UserWsMap, req.UserID)
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
			delete(UserWsMap, req.UserID)
		}

		UserWsMap[req.UserID] = userWsInfo
		//logx.Infof("【连接建立】用户 %d 已存入 UserWsMap", req.UserID)
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
		//}
		for {
			//消息类型，消息，错误
			_, p, err := conn.ReadMessage()
			if err != nil {
				//用户断开聊天
				fmt.Println(err)
				break
			}
			fmt.Println(string(p), req.UserID)
			//发送消息
			conn.WriteMessage(websocket.TextMessage, []byte("xxx"))
		}
	}
}
