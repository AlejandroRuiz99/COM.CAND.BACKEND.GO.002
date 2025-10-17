package nats

import (
	"context"
	"fmt"
	"time"

	natslib "github.com/nats-io/nats.go"
)

// Client es un wrapper del cliente NATS con configuración optimizada para IoT
type Client struct {
	nc *natslib.Conn
}

// NewClient crea un nuevo cliente NATS conectado al servidor especificado
func NewClient(url string) (*Client, error) {
	opts := []natslib.Option{
		natslib.Name("iot-sensor-client"),
		natslib.Timeout(10 * time.Second),
		natslib.ReconnectWait(2 * time.Second),
		natslib.MaxReconnects(10),
		natslib.DisconnectErrHandler(func(nc *natslib.Conn, err error) {
			// Log de desconexión (en producción usaríamos logger estructurado)
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		natslib.ReconnectHandler(func(nc *natslib.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
	}

	nc, err := natslib.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &Client{nc: nc}, nil
}

// IsConnected verifica si el cliente está conectado al servidor NATS
func (c *Client) IsConnected() bool {
	return c.nc != nil && c.nc.IsConnected()
}

// Publish publica un mensaje en un subject
func (c *Client) Publish(subject string, data []byte) error {
	if err := c.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, err)
	}
	return nil
}

// Subscribe crea una suscripción a un subject con un handler
func (c *Client) Subscribe(subject string, handler natslib.MsgHandler) (*natslib.Subscription, error) {
	sub, err := c.nc.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	return sub, nil
}

// Request envía un request y espera una respuesta con timeout controlado por context
func (c *Client) Request(ctx context.Context, subject string, data []byte) (*natslib.Msg, error) {
	msg, err := c.nc.RequestWithContext(ctx, subject, data)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", subject, err)
	}
	return msg, nil
}

// Close cierra la conexión NATS de forma limpia (drain + close)
func (c *Client) Close() error {
	if c.nc != nil {
		// Drain espera a que se envíen mensajes pendientes antes de cerrar
		c.nc.Drain()
		// Close cierra definitivamente la conexión
		c.nc.Close()
	}
	return nil
}
