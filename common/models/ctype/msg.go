package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type MsgType int8

const (
	TextMsgType MsgType = iota + 1
	ImageMsgType
	VideoMsgType
	FileMsgType
	VoiceMsgType
	VoiceTelMsgType
	VideoTelMsgType
	RecallMsgType
	ReplyMsgType
	QuoteMsgType
	AtMsgType
	TipMsgType
)

type SystemMsg struct {
	Type int8 `json:"type"` //违规类型 1：涉黄 2：涉政 3：不正当言论
}

func (c *SystemMsg) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), c)
}
func (c SystemMsg) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

type Msg struct {
	Type        MsgType      `json:"type"`
	Content     *string      `json:"content"`
	TextMsg     *TextMsg     `json:"text_msg,omitempty"`
	ImageMsg    *ImageMsg    `json:"image_msg,omitempty"`
	VideoMsg    *VideoMsg    `json:"video_msg,omitempty"`
	FileMsg     *FileMsg     `json:"file_msg,omitempty"`
	VoiceMsg    *VoiceMsg    `json:"voice_msg,omitempty"`
	VoiceTelMsg *VoiceTelMsg `json:"voice_tel_msg,omitempty"`
	VideoTelMsg *VideoTelMsg `json:"video_tel_msg,omitempty"`
	RecallMsg   *RecallMsg   `json:"recall_msg,omitempty"`
	ReplyMsg    *ReplyMsg    `json:"reply_msg,omitempty"`
	QuoteMsg    *QuoteMsg    `json:"quote_msg,omitempty"`
	AtMsg       *AtMsg       `json:"at_msg,omitempty"`  //@消息（群聊才有）
	TipMsg      *TipMsg      `json:"tip_msg,omitempty"` //一般是不入库的
}

func (c *Msg) Scan(value interface{}) error {
	//err := json.Unmarshal(value.([]byte), c)
	//if err != nil {
	//	return err
	//}
	//if c.Type == RecallMsgType {
	//	//如果这个消息是撤回消息，那就不要把原消息传回
	//	if c.RecallMsg != nil {
	//		c.RecallMsg.OriginMsg = nil
	//	}
	//}
	return json.Unmarshal(value.([]byte), c)
}
func (c Msg) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

type TextMsg struct {
	Content string `json:"content"`
}
type ImageMsg struct {
	Title string `gorm:"size:256" json:"title"`
	Url   string `gorm:"size:256" json:"url"`
}
type VideoMsg struct {
	Title string `gorm:"size:32" json:"title"`
	Url   string `gorm:"size:64" json:"url"`
	Time  int    `json:"time"`
}
type FileMsg struct {
	Title string `gorm:"size:32" json:"title"`
	Url   string `gorm:"size:64" json:"url"`
	Size  int64  `json:"size"`
	Type  string `gorm:"size:32" json:"type"`
}
type VoiceMsg struct {
	Url  string `gorm:"size:64" json:"url"`
	Time int    `json:"time"`
}
type VoiceTelMsg struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	EndReason string    `json:"end_reason"` //0:发起方挂断 1：接收方挂断  2：网络原因挂断  3：未打通
}
type VideoTelMsg struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	EndReason string    `json:"end_reason"` //0:发起方挂断 1：接收方挂断  2：网络原因挂断  3：未打通
}
type RecallMsg struct {
	Notice    string `gorm:"size:64" json:"notice"` //撤回的提示词
	MsgID     uint   `json:"msg_id"`                //需要撤回的消息ID *入参必填*
	OriginMsg *Msg   `json:"origin_msg,omitempty"`  //源消息
}
type ReplyMsg struct {
	MsgId         uint      `json:"msg_id"`
	Content       string    `gorm:"size:256" json:"content"`
	MsgContent    *Msg      `json:"msg_content,omitempty"`
	UserId        uint      `json:"user_id"`         //被回复人的ID
	UserNickName  string    `json:"user_nick_name"`  //被回复人的昵称
	OriginMsgDate time.Time `json:"origin_msg_date"` //被回复消息的时间
}

type QuoteMsg struct {
	MsgId         uint      `json:"msg_id"`
	Content       string    `gorm:"size:256" json:"content"`
	MsgContent    *Msg      `json:"msg_content"`
	UserId        uint      `json:"user_id"`         //被引用人的ID
	UserNickName  string    `json:"user_nick_name"`  //被引用人的昵称
	OriginMsgDate time.Time `json:"origin_msg_date"` //被引用消息的时间
}
type AtMsg struct {
	UserId     uint   `json:"user_id"`
	Content    string `gorm:"size:256" json:"content"`
	MsgContent *Msg   `json:"msg_content"`
}
type TipMsg struct {
	Status  string `gorm:"size:32" json:"status"` //error warning success info
	Content string `gorm:"size:256" json:"content"`
}
