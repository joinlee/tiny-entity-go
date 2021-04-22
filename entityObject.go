package tiny

type Entity interface {
	//获取表名称（表名称和实体名称保持一致）
	TableName() string
}

type IEntityObject interface {
	IQueryObject
	ITakeChildQueryObject

	//获取表名称（表名称和实体名称保持一致）
	TableName() string
	// ToString() string
}

type IQueryObject interface {
	IResultQueryObject
	// 添加查询条件
	// queryStr 查询语句， args 条件参数
	// 例： ctx.User.Where("Id = ?", user.Id).Any()
	Where(queryStr interface{}, args ...interface{}) IQueryObject
	//添加指定表的查询条件
	/* entity 需要查询的实体 queryStr 查询语句， args 条件参数
	   entity 表示查询外键表的条件
	   ex： ctx.User.WhereWith(ctx.Account, "Id = ?", user.Id).Any() */
	WhereWith(entity Entity, queryStr interface{}, args ...interface{}) IQueryObject
	//对指定字段进行顺序排序
	OrderBy(fields interface{}) IQueryObject
	//对指定字段进行倒序排序
	OrderByDesc(fields interface{}) IQueryObject
	IndexOf() IQueryObject

	//对指定字段进行分组
	GroupBy(fields interface{}) IResultQueryObject
	Select(fields ...interface{}) IResultQueryObject
	//查询字段包含的值
	Contains(field string, values interface{}) IResultQueryObject

	// 需要获取的数据行数
	Take(count int) ITakeChildQueryObject
	// 添加外联引用
	/*
		fEntity 需要连接的实体， mField 主表的连接字段， fField 外联表的字段
	*/
	JoinOn(foreignEntity Entity, mField string, fField string) IQueryObject
}

type IResultQueryObject interface {
	IAssembleResultQuery
	Max() float64
	Min() float64
	Count() int
	// 查询数据库，并返回是否存在结果
	Any() bool
	// 查询数据库，并返回第一个结果
	/*
		args 需要返回的实体的指针
		return （ isNil,  *Empty）
		例子：
		user := models.User{}
		ctx.User.First(&user)
	*/
	First(interface{}) (bool, *Empty)
}

type ITakeChildQueryObject interface {
	IResultQueryObject
	// 需要跳过的数据行数
	Skip(count int) IAssembleResultQuery
}

type IAssembleResultQuery interface {
	// 查询数据库，并返回列表结果
	/*
		args 需要返回的实体的指针
		例子：
		list := make([]models.User, 0)
		ctx.User.ToList(&list)
	*/
	ToList(interface{})
}

type Empty struct{}
