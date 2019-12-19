package models

type BookCategory struct {
	Id         int // 自增主键
	BookId     int
	CategoryId int
}

func (m *BookCategory) TableName() string {
	return TNBookCategory()
}
