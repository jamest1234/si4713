package si4713

const (
	CmdPowerUp       = 0x1
	CmdGetRev        = 0x10
	CmdSetProperty   = 0x12
	CmdGetIntStatus  = 0x14
	CmdSetTxFreq     = 0x30
	CmdSetTxPower    = 0x31
	CmdTxTuneMeasure = 0x32
	CmdTxTuneStatus  = 0x33
	CmdTxASQStatus   = 0x34
	CmdTxRDSBuff     = 0x35
	CmdTxRDSPs       = 0x36
	CmdGPOCtl        = 0x80
	CmdGPOSet        = 0x81
)

const (
	PropRefClkFreq           = 0x201
	PropTxComponentEnable    = 0x2100
	PropTxAudioDeviation     = 0x2101
	PropTxRDSDeviation       = 0x2103
	PropTxPreemphasis        = 0x2106
	PropTxACompEnable        = 0x2200
	PropTxACompThreshold     = 0x2201
	PropTxACompAttackTime    = 0x2202
	PropTxACompReleaseTime   = 0x2203
	PropTxACompGain          = 0x2204
	PropTxRDSInterruptSource = 0x2C00
	PropTxRDSPI              = 0x2C01
	PropTxRDSPsMix           = 0x2C02
	PropTxRDSPsMisc          = 0x2C03
	PropTxRDSPsRepeatCount   = 0x2C04
	PropTxRDSMessageCount    = 0x2C05
	PropTxRDSPsAF            = 0x2C06
	PropTxRDSFIFOSize        = 0x2C07
)

const (
	DynamicRangeControl = 1 << iota
	AudioLimiter
)

type AudioCompressionSettings struct {
	Enable      uint16
	Threshold   int     // Range from -40dB to 0dB
	AttackTime  float64 // Range from 0.5ms to 5ms
	ReleaseTime int     // 100ms, 200ms, 350ms, 525ms, or 1000ms
	Gain        uint16  // Range from 0dB to 20dB
}

var MinimalCompression = AudioCompressionSettings{
	Enable:      DynamicRangeControl | AudioLimiter,
	Threshold:   -40,
	AttackTime:  5,
	ReleaseTime: 100,
	Gain:        15,
}

var AggressiveCompression = AudioCompressionSettings{
	Enable:      DynamicRangeControl | AudioLimiter,
	Threshold:   -15,
	AttackTime:  0.5,
	ReleaseTime: 1000,
	Gain:        5,
}
