package dao

// User 示例用户模型
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:255;not null;unique"`
	Email string `gorm:"size:255;not null;unique"`
}
