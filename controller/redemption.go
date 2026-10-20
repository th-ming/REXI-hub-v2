// Copyright (c) 2026 REXI-hub team
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package controller

import (
	"net/http"
	"strconv"
	"rexi-hub-v2/common"
	"rexi-hub-v2/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllRedemptions(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 0 {
		p = 0
	}
	if pageSize < 1 {
		pageSize = common.ItemsPerPage
	}
	redemptions, total, err := model.GetAllRedemptions((p-1)*pageSize, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     redemptions,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func SearchRedemptions(c *gin.Context) {
	keyword := c.Query("keyword")
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 0 {
		p = 0
	}
	if pageSize < 1 {
		pageSize = common.ItemsPerPage
	}
	redemptions, total, err := model.SearchRedemptions(keyword, (p-1)*pageSize, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     redemptions,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func GetRedemption(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	redemption, err := model.GetRedemptionById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    redemption,
	})
	return
}

func AddRedemption(c *gin.Context) {
	redemption := model.Redemption{}
	err := c.ShouldBindJSON(&redemption)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if len(redemption.Name) == 0 || len(redemption.Name) > 20 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Redemption code name must be 1-20 characters long",
		})
		return
	}

	// Kiểm tra khoảng thời gian
	if redemption.ValidFrom > 0 && redemption.ValidUntil > 0 && redemption.ValidFrom >= redemption.ValidUntil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Effective time must be earlier than expiry time",
		})
		return
	}

	var keys []string
	if redemption.Key != "" {
		// If key is provided, use it and check for duplicates
		_, err := model.GetRedemptionByKey(redemption.Key)
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		if err == nil { // If err is nil, a record was found
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Redemption code content already exists",
			})
			return
		}

		cleanRedemption := model.Redemption{
			UserId:      c.GetInt("id"),
			Name:        redemption.Name,
			Key:         redemption.Key,
			CreatedTime: common.GetTimestamp(),
			Quota:       redemption.Quota,
			IsGift:      redemption.IsGift,
			MaxUses:     redemption.MaxUses,
			ValidFrom:   redemption.ValidFrom,
			ValidUntil:  redemption.ValidUntil,
		}
		err = cleanRedemption.Insert()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		keys = append(keys, redemption.Key)

	} else {
		// If key is not provided, generate multiple random keys
		if redemption.Count <= 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Redemption code count must be greater than 0",
			})
			return
		}
		if redemption.Count > 100 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Batch redemption code generation cannot exceed 100 at a time",
			})
			return
		}

		for i := 0; i < redemption.Count; i++ {
			key := common.GetUUID()
			cleanRedemption := model.Redemption{
				UserId:      c.GetInt("id"),
				Name:        redemption.Name,
				Key:         key,
				CreatedTime: common.GetTimestamp(),
				Quota:       redemption.Quota,
				IsGift:      redemption.IsGift,
				MaxUses:     redemption.MaxUses,
				ValidFrom:   redemption.ValidFrom,
				ValidUntil:  redemption.ValidUntil,
			}
			err = cleanRedemption.Insert()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
					"data":    keys,
				})
				return
			}
			keys = append(keys, key)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
	return
}

func DeleteRedemption(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := model.DeleteRedemptionById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateRedemption(c *gin.Context) {
	statusOnly := c.Query("status_only")
	redemption := model.Redemption{}
	err := c.ShouldBindJSON(&redemption)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	cleanRedemption, err := model.GetRedemptionById(redemption.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if statusOnly != "" {
		cleanRedemption.Status = redemption.Status
	} else {
		// If you add more fields, please also update redemption.Update()
		cleanRedemption.Name = redemption.Name
		cleanRedemption.Quota = redemption.Quota
		cleanRedemption.ValidFrom = redemption.ValidFrom
		cleanRedemption.ValidUntil = redemption.ValidUntil
	}
	err = cleanRedemption.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanRedemption,
	})
	return
}

// CountRedemptionsByName đếm số lượng mã đổi thưởng theo tên
func CountRedemptionsByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Name cannot be empty",
		})
		return
	}

	count, err := model.CountRedemptionsByName(name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    count,
	})
}

// DeleteRedemptionsByName xóa hàng loạt mã đổi thưởng theo tên
func DeleteRedemptionsByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Name cannot be empty",
		})
		return
	}

	count, err := model.DeleteRedemptionsByName(name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    count,
	})
}

// BatchDisableRedemptions vô hiệu hóa hàng loạt mã đổi thưởng
func BatchDisableRedemptions(c *gin.Context) {
	var requestData struct {
		Ids []int `json:"ids"`
	}

	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if len(requestData.Ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "ID list cannot be empty",
		})
		return
	}

	count, err := model.BatchDisableRedemptions(requestData.Ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    count,
	})
}

// DeleteDisabledRedemptions xóa toàn bộ mã đổi thưởng đã bị vô hiệu hóa
func DeleteDisabledRedemptions(c *gin.Context) {
	count, err := model.DeleteDisabledRedemptions()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    count,
	})
}
