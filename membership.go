package main

type membership struct {
	Type int `json:"type"`
	Cost int `json:"cost"`
}

const (
	MEMBERSHIP_TYPE_NO = 0
	MEMBERSHIP_COST_NO = 0

	MEMBERSHIP_NETLIFX_TYPE_BASIC = iota - 1 // 1
	MEMBERSHIP_NETLIFX_TYPE_STANDARD
	MEMBERSHIP_NETLIFX_TYPE_PREMIUM

	MEMBERSHIP_WAVVE_TYPE_BASIC
	MEMBERSHIP_WAVVE_TYPE_STANDARD
	MEMBERSHIP_WAVVE_TYPE_PREMIUM
	MEMBERSHIP_WAVVE_TYPE_FLO
	MEMBERSHIP_WAVVE_TYPE_BUGS
	MEMBERSHIP_WAVVE_TYPE_KB

	MEMBERSHIP_NETLIFX_COST_STANDARD = 9_500
	MEMBERSHIP_NETLIFX_COST_BASIC    = 13_500
	MEMBERSHIP_NETLIFX_COST_PREMIUM  = 17_000

	MEMBERSHIP_WAVVE_COST_BASIC    = 7_900
	MEMBERSHIP_WAVVE_COST_STANDARD = 10_900
	MEMBERSHIP_WAVVE_COST_PREMIUM  = 13_900
	MEMBERSHIP_WAVVE_COST_FLO      = 13_750
	MEMBERSHIP_WAVVE_COST_BUGS     = 13_750
	MEMBERSHIP_WAVVE_COST_KB       = 6_700
)
