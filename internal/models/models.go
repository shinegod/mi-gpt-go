package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Config 系统配置模型（存储在数据库中）
type Config struct {
	ID    int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Key   string `gorm:"uniqueIndex;not null" json:"key"`   // 配置键，如 "ai.provider", "speaker.userID"
	Value string `gorm:"type:text" json:"value"`           // 配置值（JSON格式）
	Type  string `gorm:"not null" json:"type"`             // 配置类型：string, int, bool, array, object
	Group string `gorm:"not null" json:"group"`            // 配置分组：ai, speaker, bot, database
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// User 用户模型
type User struct {
	ID                  string            `gorm:"type:char(36);primaryKey" json:"id"`
	Name                string            `gorm:"not null" json:"name"`
	Profile             string            `gorm:"not null" json:"profile"`
	Rooms               []Room            `gorm:"many2many:room_members;" json:"rooms,omitempty"`
	Messages            []Message         `gorm:"foreignKey:SenderID" json:"messages,omitempty"`
	Memories            []Memory          `gorm:"foreignKey:OwnerID" json:"memories,omitempty"`
	ShortTermMemories   []ShortTermMemory `gorm:"foreignKey:OwnerID" json:"shortTermMemories,omitempty"`
	LongTermMemories    []LongTermMemory  `gorm:"foreignKey:OwnerID" json:"longTermMemories,omitempty"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// BeforeCreate GORM hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// Room 房间模型
type Room struct {
	ID                  string            `gorm:"type:char(36);primaryKey" json:"id"`
	Name                string            `gorm:"not null" json:"name"`
	Description         string            `gorm:"not null" json:"description"`
	Members             []User            `gorm:"many2many:room_members;" json:"members,omitempty"`
	Messages            []Message         `gorm:"foreignKey:RoomID" json:"messages,omitempty"`
	Memories            []Memory          `gorm:"foreignKey:RoomID" json:"memories,omitempty"`
	ShortTermMemories   []ShortTermMemory `gorm:"foreignKey:RoomID" json:"shortTermMemories,omitempty"`
	LongTermMemories    []LongTermMemory  `gorm:"foreignKey:RoomID" json:"longTermMemories,omitempty"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// BeforeCreate GORM hook
func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// Message 消息模型
type Message struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Text      string    `gorm:"not null" json:"text"`
	SenderID  string    `gorm:"type:char(36);not null" json:"senderId"`
	Sender    User      `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	RoomID    string    `gorm:"type:char(36);not null" json:"roomId"`
	Room      Room      `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Memories  []Memory  `gorm:"foreignKey:MessageID" json:"memories,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Memory 记忆模型
type Memory struct {
	ID                  int               `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID           int               `gorm:"not null" json:"messageId"`
	Message             Message           `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	OwnerID             *string           `gorm:"type:char(36)" json:"ownerId"`
	Owner               *User             `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	RoomID              string            `gorm:"type:char(36);not null" json:"roomId"`
	Room                Room              `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	ShortTermMemories   []ShortTermMemory `gorm:"foreignKey:CursorID" json:"shortTermMemories,omitempty"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// ShortTermMemory 短期记忆模型
type ShortTermMemory struct {
	ID                int              `gorm:"primaryKey;autoIncrement" json:"id"`
	Text              string           `gorm:"not null" json:"text"`
	CursorID          int              `gorm:"not null" json:"cursorId"`
	Cursor            Memory           `gorm:"foreignKey:CursorID" json:"cursor,omitempty"`
	OwnerID           *string          `gorm:"type:char(36)" json:"ownerId"`
	Owner             *User            `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	RoomID            string           `gorm:"type:char(36);not null" json:"roomId"`
	Room              Room             `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	LongTermMemories  []LongTermMemory `gorm:"foreignKey:CursorID" json:"longTermMemories,omitempty"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

// LongTermMemory 长期记忆模型
type LongTermMemory struct {
	ID        int             `gorm:"primaryKey;autoIncrement" json:"id"`
	Text      string          `gorm:"not null" json:"text"`
	CursorID  int             `gorm:"not null" json:"cursorId"`
	Cursor    ShortTermMemory `gorm:"foreignKey:CursorID" json:"cursor,omitempty"`
	OwnerID   *string         `gorm:"type:char(36)" json:"ownerId"`
	Owner     *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	RoomID    string          `gorm:"type:char(36);not null" json:"roomId"`
	Room      Room            `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
} 