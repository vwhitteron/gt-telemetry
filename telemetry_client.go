package telemetry

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"time"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/rs/zerolog"
	"github.com/vwhitteron/gt-telemetry/internal/gttelemetry"
	"github.com/vwhitteron/gt-telemetry/internal/vehicles"
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

type GTClientOpts struct {
	IPAddr       string
	LogLevel     string
	Logger       *zerolog.Logger
	StatsEnabled bool
	VehicleDB    string
}

type GTClient struct {
	log         zerolog.Logger
	ipAddr      string
	sendPort    int
	receivePort int
	Telemetry   *transformer
	Statistics  *statistics
}

func NewGTClient(opts GTClientOpts) (*GTClient, error) {
	var log zerolog.Logger
	if opts.Logger != nil {
		log = *opts.Logger
	} else {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	switch opts.LogLevel {
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "info":
		log = log.Level(zerolog.InfoLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	default:
		log = log.Level(zerolog.WarnLevel)
		log.Warn().Msg("Invalid log level, defaulting to warn")
	}

	if opts.IPAddr == "" {
		opts.IPAddr = "255.255.255.255"
	}

	inventory, err := vehicles.NewInventory(opts.VehicleDB)
	if err != nil {
		return nil, err
	}

	return &GTClient{
		log:         log,
		ipAddr:      opts.IPAddr,
		sendPort:    33739,
		receivePort: 33740,
		Telemetry:   NewTransformer(inventory),
		Statistics: &statistics{
			enabled:           opts.StatsEnabled,
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
		c.log.Fatal().Msgf("resolve UDP address: %s", err.Error())
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		c.log.Fatal().Msgf("listen UDP: %s", err.Error())
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
			c.log.Debug().Msgf("failed to receive telemetry: %s", err.Error())
			continue
		}

		if len(buffer[:bufLen]) == 0 {
			c.log.Debug().Msg("no data received")
			continue
		}

		decodeStart := time.Now()

		telemetryData := salsa20Decode(buffer[:bufLen])

		reader := bytes.NewReader(telemetryData)
		stream := kaitai.NewStream(reader)

		err = rawTelemetry.Read(stream, nil, nil)
		if err != nil {
			c.log.Error().Msgf("failed to parse telemetry: %s", err.Error())
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
			c.log.Warn().Int("dropped", delta-1).Msg("dropped packets detected")
			c.Statistics.PacketsDropped += int(delta - 1)
		} else if delta < 0 {
			c.log.Warn().Msg("delayed packet detected")
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
	c.log.Debug().Msg("Sending heartbeat")
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
		c.log.Fatal().Msgf("write to udp: %s", err.Error())
	}
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		c.log.Fatal().Msgf("set read deadline: %s", err.Error())
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
