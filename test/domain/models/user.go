/*
 * @Author: john lee
 * @Date: 2022-03-30 09:22:46
 * @LastEditors: john lee
 * @LastEditTime: 2022-03-30 10:16:30
 * @FilePath: \tiny-entity-go\test\domain\models\user.go
 * @Description:
 *
 * Copyright (c) 2022 by 用户/公司名, All Rights Reserved.
 */
package models

import "github.com/joinlee/tiny-entity-go"

type User struct {
	tiny.IEntityObject[User]
	Id        string   `tiny:"primaryKey" json:"id"`
	Name      string   `tiny:"varchar(255);notNull;index" json:"name"`
	Phone     string   `tiny:"varchar(255);index" json:"phone"`
	AccountId string   `tiny:"varchar(32);index" json:"accountId"`
	Account   *Account `tiny:"mapping:Account" json:"account"`
}
