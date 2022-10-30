package yafw

type Interface struct {
	Name string `json:"name"`
	MTU  int    `json:"mtu"`
	MAC  string `json:"mac"`
	Up   bool   `json:"up"`
	Zone string `json:"zone"`
}
