package block

type divider struct {}

func (divider) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"divider"}`), nil
}
func (divider) blockAble(){}

func Divider() divider {
	return divider{}
}
