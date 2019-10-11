package lists

// MysqlKeyword converts a SortDirection object to a string understood by mysql
//   e.g. {Name: "tenant_id", Direction: Direction_Descending }
func (sd Direction) ToMysqlKeyword() string {
	switch sd {
	case Direction_Descending:
		return "DESC"
	}

	return "ASC"
}
