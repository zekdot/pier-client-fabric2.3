package main

type Fabric struct {
	Name        string `toml:"name" json:"name"`
	Addr        string `toml:"addr" json:"addr"`
	OrganizationsPath        string `toml:"organizations_path" json:"organizations_path"`
	Username    string `toml:"username" json:"username"`
	CCID        string `toml:"ccid" json:"ccid"`
	ChannelId   string `mapstructure:"channel_id" toml:"channel_id" json:"channel_id"`

}

func DefaultConfig() *Fabric {
	return &Fabric{
		Addr:        "40.125.164.122:10053",
		Name:        "fabric2.3",
		OrganizationsPath: ".",
		Username:    "Admin",
		CCID:        "Broker-001",
		ChannelId:   "mychannel",
	}
}