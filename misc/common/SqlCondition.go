package common

import "fmt"

type SQLConditions struct{
	conditions []string
}

func (this *SQLConditions) AddCondition(cond string){
	this.conditions = append(this.conditions, cond)
}

func (this *SQLConditions) GetSql() (ret string){
	if len(this.conditions) == 0{
		return
	}else{
		ret = " WHERE "
		for k, v := range this.conditions{
			if k == 0{
				ret += v
			}else {
				ret += fmt.Sprintf(" AND %s", v)
			}
		}
	}
	return
}
