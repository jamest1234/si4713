package si4713

const (
	CmdPowerUp      = 0x1
	CmdGetRev       = 0x10
	CmdPowerDown    = 0x11
	CmdSetProperty  = 0x12
	CmdGetIntStatus = 0x14
	CmdSetTxFreq    = 0x30
	CmdSetTxPower   = 0x31
)

const (
	PropRefClkFreq    = 0x201
	PropTxPreemphasis = 0x2106
	PropTxACompGain   = 0x2204
	PropTxACompEnable = 0x2200
)
