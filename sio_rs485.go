package sio

const (
    SER_RS485_ENABLED = 0x01
    SER_RS485_RTS_ON_SEND = 0x02
    SER_RS485_RTS_AFTER_SEND = 0x04
    SER_RS485_RX_DURING_TX = 0x10
)

type Rs485 struct {
	Enabled bool
	Loopback bool
	Rts_level_for_tx, Rts_level_for_rx bool
	Delay_before_tx, Delay_before_rx bool
	Delay_before_tx_value, Delay_before_rx_value float64
}

func (self *Rs485) Update(buf *[8]uint32) {
	if self.Enabled {
		buf[0] |= SER_RS485_ENABLED
		if self.Loopback {
			buf[0] |= SER_RS485_RX_DURING_TX
		} else {
			buf[0] &= ^uint32(SER_RS485_RX_DURING_TX)
		}
		if self.Rts_level_for_tx {
			buf[0] |= SER_RS485_RTS_ON_SEND
		} else {
			buf[0] &= ^uint32(SER_RS485_RTS_ON_SEND)
		}
		if self.Rts_level_for_rx {
			buf[0] |= SER_RS485_RTS_AFTER_SEND
		} else {
			buf[0] &= ^uint32(SER_RS485_RTS_AFTER_SEND)
		}
		if self.Delay_before_tx {
			buf[1] = uint32(self.Delay_before_tx_value * 1000.)
		}
		if self.Delay_before_rx {
			buf[2] = uint32(self.Delay_before_rx_value * 1000.)
		}
	} else {
		buf[0] = 0
	}
}

/* EOF */
