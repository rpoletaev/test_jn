package test_jn

func lpush(l *[]interface{}, newList []interface{}) {
	*l = append(newList, *l...)
}

func rpush(l *[]interface{}, newList []interface{}) {
	*l = append(*l, newList...)
}

func lpop(l *[]interface{}) interface{} {
	item := (*l)[0]
	newList := make([]interface{}, len(*l)-1)
	copy(newList, (*l)[1:])
	*l = newList
	return item
}

func rpop(l *[]interface{}) interface{} {
	item := (*l)[len(*l)-1]
	*l = (*l)[0 : len(*l)-1]
	return item
}
