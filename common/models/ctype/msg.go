package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"time"
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
	Type        int          `json:"type"`
	Content     *string      `json:"content"`
	ImageMsg    *ImageMsg    `json:"image_msg"`
	VideoMsg    *VideoMsg    `json:"video_msg"`
	FileMsg     *FileMsg     `json:"file_msg"`
	VoiceMsg    *VoiceMsg    `json:"voice_msg"`
	VoiceTelMsg *VoiceTelMsg `json:"voice_tel_msg"`
	VideoTelMsg *VideoTelMsg `json:"video_tel_msg"`
	RecallMsg   *RecallMsg   `json:"recall_msg"`
	ReplyMsg    *ReplyMsg    `json:"reply_msg"`
	QuoteMsg    *QuoteMsg    `json:"quote_msg"`
	AtMsg       *AtMsg       `json:"at_msg"` //@消息（群聊才有）
}

func (c *Msg) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), c)
}
func (c Msg) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	return string(b), err
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
	OriginMsg *Msg   `json:"origin_msg"`            //源消息
}
type ReplyMsg struct {
	MsgId      uint   `json:"msg_id"`
	Content    string `gorm:"size:256" json:"content"`
	MsgContent *Msg   `json:"msg_content"`
}

type QuoteMsg struct {
	MsgId      uint   `json:"msg_id"`
	Content    string `gorm:"size:256" json:"content"`
	MsgContent *Msg   `json:"msg_content"`
}
type AtMsg struct {
	UserId     uint   `json:"user_id"`
	Content    string `gorm:"size:256" json:"content"`
	MsgContent *Msg   `json:"msg_content"`
}
