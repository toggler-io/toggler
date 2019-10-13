package release

type IPAllow struct {
	ID     string `ext:"ID"`
	FlagID string
	InternetProtocolAddress string
}

func (a IPAllow) CheckIP(ipAddr string) (value bool, ok bool) {
	if a.InternetProtocolAddress == `` {
		return false, false
	}
	return a.InternetProtocolAddress == ipAddr, true
}