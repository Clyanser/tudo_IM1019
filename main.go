package main

import (
	"flag"
	"fmt"
	"tudo_IM1019/core"
	"tudo_IM1019/tudoIM_chat/chat_models"
	"tudo_IM1019/tudoIM_file/file_models"
	"tudo_IM1019/tudoIM_group/group_models"
	"tudo_IM1019/tudoIM_user/user_models"
)

type Options struct {
	DB bool
}

func main() {

	var opt Options
	flag.BoolVar(&opt.DB, "db", false, "db")
	flag.Parse()

	if opt.DB {
		db := core.InitMysql()
		err := db.AutoMigrate(
			&user_models.UserModel{},           // 用户表
			&user_models.FriendModel{},         // 好友表
			&user_models.FriendVerifyModel{},   // 好友验证表
			&user_models.UserConfModel{},       // 用户配置表
			&chat_models.ChatModel{},           // 对话表
			&chat_models.TopUserModel{},        // 置顶用户表
			&chat_models.UserChatDeleteModel{}, //用户删除聊天记录表
			&group_models.GroupModel{},         // 群组表
			&group_models.GroupMemberModel{},   // 群成员表
			&group_models.GroupMsgModel{},      // 群消息表
			&group_models.GroupVerifyModel{},   // 群验证表
			&file_models.FileModel{},           //文件表
		)
		if err != nil {
			fmt.Println("表结构生成失败", err)
			return
		}
		fmt.Println("表结构生成成功！")
	}
}
