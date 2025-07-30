package list_query

import (
	"fmt"
	"gorm.io/gorm"
	"tudo_IM1019/common/models"
)

type Option struct {
	PageInfo models.PageInfo
	Where    *gorm.DB //高级查询
	likes    []string //模糊匹配的字段
	Preload  []string //预加载字段
}

// func ListQuery[T any](db *gorm.DB, model T, option Option) (list []T, count int64, err error) {
//
//		query := db.Where(model) //把结构体自己的查询条件查了
//		//模糊匹配
//		if option.PageInfo.Key != "" && len(option.Preload) > 0 {
//			likeQuery := db.Where("")
//			for index, column := range option.likes {
//				if index == 0 {
//					likeQuery.Where(fmt.Sprintf("%s like '%%%?%%'", column), option.PageInfo.Key)
//				} else {
//					likeQuery.Or(fmt.Sprintf("%s like '%%%?%%'", column), option.PageInfo.Key)
//				}
//			}
//			query.Where(likeQuery)
//		}
//
//		//求总数
//		query.Model(model).Count(&count)
//
//		//preload（预加载） 部分
//		for _, s := range option.Preload {
//			query = query.Preload(s)
//		}
//
//		//分页查询
//		if option.PageInfo.Page <= 0 {
//			option.PageInfo.Page = 1
//		}
//		if option.PageInfo.Limit <= 0 {
//			option.PageInfo.Limit = 10
//		}
//
//		offset := option.PageInfo.Limit * (option.PageInfo.Page - 1)
//
//		err = query.Limit(option.PageInfo.Limit).Offset(offset).Find(&list).Error
//		if err != nil {
//			return
//		}
//		return
//	}
func ListQuery[T any](db *gorm.DB, model T, option Option) (list []T, count int64, err error) {
	query := db.Model(&model)

	// ✅ 应用传入的 WHERE 条件
	if option.Where != nil {
		query = query.Where(option.Where)
	}

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

	// 查询总数
	query.Count(&count)

	// 预加载
	for _, preload := range option.Preload {
		query = query.Preload(preload)
	}

	if option.PageInfo.Page <= 0 {
		option.PageInfo.Page = 1
	}
	if option.PageInfo.Limit <= 0 {
		option.PageInfo.Limit = 10
	}

	offset := option.PageInfo.Limit * (option.PageInfo.Page - 1)

	err = query.Limit(option.PageInfo.Limit).Offset(offset).Find(&list).Error
	if err != nil {
		return
	}
	return
}
