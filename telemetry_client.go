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
type Config struct {
	IPAddr       string
	LogLevel     string
}

type GTClient struct {
	logger      *utils.Logger
	ipAddr      string
	sendPort    int
	receivePort int
	Telemetry   *transformer
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

	c.SendHeartbeat(conn)

	rawTelemetry := gttelemetry.NewGranTurismoTelemetry()

	for {
		buffer := make([]byte, 4096)
		c.logger.Debug("Reading telemetry")
		bufLen, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			c.logger.Debug(fmt.Sprintf("Failed to receive telemetry: %s", err.Error()))
			c.SendHeartbeat(conn)
			continue
		}

		if len(buffer[:bufLen]) == 0 {
			c.logger.Debug("No data received")
			continue
		}

		telemetryData := salsa20Decode(buffer[:bufLen])

		reader := bytes.NewReader(telemetryData)
		stream := kaitai.NewStream(reader)

		err = rawTelemetry.Read(stream, nil, nil)
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to parse telemetry: %s", err.Error()))
		}

		c.Telemetry.RawTelemetry = *rawTelemetry
	}
}

func (c *GTClient) SendHeartbeat(conn *net.UDPConn) {
	c.logger.Debug("Sending heartbeat")

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
