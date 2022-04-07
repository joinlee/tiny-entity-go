package tiny

type Entity interface {
	//获取表名称（表名称和实体名称保持一致）
	TableName() string
}

type IEntityObject[T any] interface {
	IQueryObject[T]
	ITakeChildQueryObject[T]

	//获取表名称（表名称和实体名称保持一致）
	TableName() string
}

type IQueryObject[T any] interface {
	IResultQueryObject[T]

	// 添加where条件的AND 连接符，只能在where，In等条件查询方法之间
	// 例如：ctx.User.Where().And().Where().Tolist()
	And() IQueryObject[T]
	// 添加where条件的OR 连接符，只能在where，In等条件查询方法之间
	// 例如：ctx.User.Where().OR().Where().Tolist()
	Or() IQueryObject[T]

	// 添加查询条件
	// queryStr 查询语句， args 条件参数
	// 例： ctx.User.Where("Id = ?", user.Id).Any()
	Where(queryStr interface{}, args ...interface{}) IQueryObject[T]

	//添加指定表的查询条件
	/* entity 需要查询的实体 queryStr 查询语句， args 条件参数
	   entity 表示查询外键表的条件
	   ex： ctx.User.WhereWith(ctx.Account, "Id = ?", user.Id).Any() */
	WhereWith(entity Entity, queryStr interface{}, args ...interface{}) IQueryObject[T]
	//查询字段包含的值
	Contains(field string, values interface{}) IQueryObject[T]
	// 指定表名称的In 条件语句
	ContainsWith(entity Entity, felid string, values interface{}) IQueryObject[T]

	//对指定字段进行顺序排序
	OrderBy(fields interface{}) IQueryObject[T]
	//对指定字段进行倒序排序
	OrderByDesc(fields interface{}) IQueryObject[T]
	IndexOf() IQueryObject[T]

	//对指定字段进行分组
	GroupBy(fields interface{}) IResultQueryObject[T]
	Select(fields ...interface{}) IResultQueryObject[T]

	// 需要获取的数据行数
	Take(count int) ITakeChildQueryObject[T]
	/**
	 * @description: 添加外联引用
	 * @param {Entity} foreignEntity 外键表的实体对象
	 * @param {string} mField 要关联的主键表的字段
	 * @param {string} fField 要关联的外键表字段
	 * @return {IQueryObject[T]}
	 */
	JoinOn(foreignEntity Entity, mField string, fField string) IQueryObject[T]
	// 添加外联引用
	/*
		mEntity 主表实体 fEntity 需要连接的实体， mField 主表的连接字段， fField 外联表的字段
	*/
	JoinOnWith(mEntity Entity, fEntity Entity, mField string, fField string) IQueryObject[T]

	GetIQueryObject() IQueryObject[T]
}

type IResultQueryObject[T any] interface {
	IAssembleResultQuery[T]
	Max(field string) string
	Min(field string) string
	//查询结果数量
	Count() int
	//带有参数的Count
	// 例如： countQueryObj.CountArgs("DISTINCT(`Visit`.`Id`)")
	CountArgs(field string) int
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
	First() *T
}

type ITakeChildQueryObject[T any] interface {
	IResultQueryObject[T]
	// 需要跳过的数据行数
	Skip(count int) IAssembleResultQuery[T]
}

type IAssembleResultQuery[T any] interface {
	/**
	 * @description:  查询数据库，并返回列表结果
	 * @param {*}
	 * @return {*}
	 */
	ToList() []T
	GetIQueryObject() IQueryObject[T]
}

type Empty struct{}
