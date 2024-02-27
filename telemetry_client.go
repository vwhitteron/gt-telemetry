package telemetry_client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/vwhitteron/gt-telemetry/internal/gttelemetry"
	"github.com/vwhitteron/gt-telemetry/internal/utils"
	"golang.org/x/crypto/salsa20"
)

const cipherKey string = "Simulator Interface Packet GT7 ver 0.0"

type statistics struct {
	enabled           bool
	decodeTimeLast    time.Duration
	packetRateLast    time.Time
	packetIDLast      uint32
	DecodeTimeAvg     time.Duration
	DecodeTimeMax     time.Duration
	Heartbeat         bool
	PacketRateAvg     int
	PacketRateCurrent int
	PacketRateMax     int
	PacketsDropped    int
	PacketsInvalid    int
	PacketsTotal      int
}

type Config struct {
	IPAddr       string
	LogLevel     string
	StatsEnabled bool
}

type GTClient struct {
	logger      *utils.Logger
	ipAddr      string
	sendPort    int
	receivePort int
	Telemetry   *transformer
	Statistics  *statistics
}

func NewGTClient(config Config) (*GTClient, error) {
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	logger, err := utils.NewLogger(config.LogLevel)
	if err != nil {
		return nil, err
	}

	if config.IPAddr == "" {
		config.IPAddr = "255.255.255.255"

	}

	return &GTClient{
		logger:      logger,
		ipAddr:      config.IPAddr,
		sendPort:    33739,
		receivePort: 33740,
		Telemetry:   NewTransformer(),
		Statistics: &statistics{
			enabled:           config.StatsEnabled,
			decodeTimeLast:    time.Duration(0),
			packetRateLast:    time.Now(),
			DecodeTimeAvg:     time.Duration(0),
			DecodeTimeMax:     time.Duration(0),
			Heartbeat:         false,
			PacketRateCurrent: 0,
			PacketRateMax:     0,
			PacketRateAvg:     0,
			PacketsTotal:      0,
			PacketsDropped:    0,
			PacketsInvalid:    0,
			packetIDLast:      0,
		},
	}, nil
}

func (c *GTClient) Run() {
	addr, err := net.ResolveUDPAddr("udp", ":33740")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		c.SendHeartbeat(conn)

		for range ticker.C {
			c.SendHeartbeat(conn)
		}
	}()

	rawTelemetry := gttelemetry.NewGranTurismoTelemetry()

	for {
		buffer := make([]byte, 4096)
		bufLen, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			c.logger.Debug(fmt.Sprintf("Failed to receive telemetry: %s", err.Error()))
			continue
		}

		if len(buffer[:bufLen]) == 0 {
			c.logger.Debug("No data received")
			continue
		}

		decodeStart := time.Now()

		telemetryData := salsa20Decode(buffer[:bufLen])

		reader := bytes.NewReader(telemetryData)
		stream := kaitai.NewStream(reader)

		err = rawTelemetry.Read(stream, nil, nil)
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to parse telemetry: %s", err.Error()))
			c.Statistics.PacketsInvalid++
		}

		c.Telemetry.RawTelemetry = *rawTelemetry

		c.Statistics.decodeTimeLast = time.Since(decodeStart)
		c.collectStats()
	}
}

func (c *GTClient) collectStats() {
	if !c.Statistics.enabled {
		return
	}

	c.Statistics.PacketsTotal++

	if c.Statistics.packetIDLast != c.Telemetry.SequenceID() {
		if c.Statistics.packetIDLast == 0 {
			c.Statistics.packetIDLast = c.Telemetry.SequenceID()
			return
		}

		c.Statistics.DecodeTimeAvg = (c.Statistics.DecodeTimeAvg + c.Statistics.decodeTimeLast) / 2
		if c.Statistics.decodeTimeLast > c.Statistics.DecodeTimeMax {
			c.Statistics.DecodeTimeMax = c.Statistics.decodeTimeLast
		}

		delta := int(c.Telemetry.SequenceID() - c.Statistics.packetIDLast)
		if delta > 1 {
			c.logger.Warn(fmt.Sprintf("Dropped packets detected: %d", delta-1))
			c.Statistics.PacketsDropped += int(delta - 1)
		} else if delta < 0 {
			c.logger.Warn("Delayed packet deteted")
		}

		c.Statistics.packetIDLast = c.Telemetry.SequenceID()

		if c.Telemetry.SequenceID()%10 == 0 {
			rate := time.Since(c.Statistics.packetRateLast)
			c.Statistics.PacketRateCurrent = int(10 / rate.Seconds())
			c.Statistics.packetRateLast = time.Now()
			c.Statistics.PacketRateAvg = (c.Statistics.PacketRateAvg + c.Statistics.PacketRateCurrent) / 2
			if c.Statistics.PacketRateCurrent > c.Statistics.PacketRateMax {
				c.Statistics.PacketRateMax = c.Statistics.PacketRateCurrent
			}
		}
	}

}

func (c *GTClient) SendHeartbeat(conn *net.UDPConn) {
	c.logger.Debug("Sending heartbeat")
	c.Statistics.Heartbeat = true
	defer func() {
		time.Sleep(250 * time.Millisecond)
		c.Statistics.Heartbeat = false
	}()

	_, err := conn.WriteToUDP([]byte("A"), &net.UDPAddr{
		IP:   net.ParseIP(c.ipAddr),
		Port: c.sendPort,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Fatal(err)
	}
}

func salsa20Decode(dat []byte) []byte {
	key := [32]byte{}
	copy(key[:], cipherKey)

	nonce := make([]byte, 8)
	iv := binary.LittleEndian.Uint32(dat[0x40:0x44])
	binary.LittleEndian.PutUint32(nonce, iv^0xDEADBEAF)
	binary.LittleEndian.PutUint32(nonce[4:], iv)

	ddata := make([]byte, len(dat))
	salsa20.XORKeyStream(ddata, dat, nonce, &key)
	magic := binary.LittleEndian.Uint32(ddata[:4])
	if magic != 0x47375330 {
		return nil
	}
	return ddata
}
