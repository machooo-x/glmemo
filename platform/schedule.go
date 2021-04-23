package platform

import "glmemo/helper/container"

// ToDoList ...存储待办相关信息
var ToDoList *container.BeeMap

func init() {
	ToDoList = container.NewBeeMap()
}
