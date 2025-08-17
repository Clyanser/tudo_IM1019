package list_query

import (
	"fmt"
	"gorm.io/gorm"
	"tudo_IM1019/common/models"
)

type Option struct {
	PageInfo models.PageInfo
	Join     string
	Where    *gorm.DB             //高级查询
	likes    []string             //模糊匹配的字段
	Preload  []string             //预加载字段
	Table    func() (string, any) //子查询
	Groups   []string             //分组
	Debug    bool
}

func ListQuery[T any](db *gorm.DB, model T, option Option) (list []T, count int64, err error) {
	if option.Debug {
		db = db.Debug()
	}
	//把结构体自己的查询条件查了
	query := db.Model(&model)
	// 模糊匹配（如果需要）
	if option.PageInfo.Key != "" && len(option.likes) > 0 {
		likeQuery := db.Where("")
		for index, column := range option.likes {
			clause := fmt.Sprintf("%s LIKE ?", column)
			if index == 0 {
				likeQuery.Where(clause, "%"+option.PageInfo.Key+"%")
			} else {
				likeQuery.Or(clause, "%"+option.PageInfo.Key+"%")
			}
		}
		query.Where(likeQuery)
	}
	if option.Table != nil {
		table, data := option.Table()
		query = query.Table(table, data)
	}

	if option.Join != "" {
		query = query.Joins(option.Join)
	}
	// ✅ 应用传入的 WHERE 条件
	if option.Where != nil {
		query = query.Where(option.Where)
	}
	if len(option.Groups) > 0 {
		for _, group := range option.Groups {
			query = query.Group(group)
		}
	}

	// 查询总数
	query.Count(&count)

	// 预加载
	for _, preload := range option.Preload {
		query = query.Preload(preload)
	}

	if option.PageInfo.Page <= 0 {
		option.PageInfo.Page = 1
	}

	if option.PageInfo.Limit != -1 { //如果是-1 就是查全部
		if option.PageInfo.Limit <= 0 {
			option.PageInfo.Limit = 10
		}
	}

	offset := option.PageInfo.Limit * (option.PageInfo.Page - 1)
	if option.PageInfo.Sort != "" {
		query = query.Order(option.PageInfo.Sort)
	}
	err = query.Limit(option.PageInfo.Limit).Offset(offset).Find(&list).Error
	if err != nil {
		return
	}
	return
}
