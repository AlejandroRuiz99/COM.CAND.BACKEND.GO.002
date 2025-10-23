package nats

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natslib "github.com/nats-io/nats.go"
)

// setupTestNATS crea un servidor NATS en memoria para testing
func setupTestNATS(t *testing.T) (*natsserver.Server, string) {
	t.Helper()

	opts := &natsserver.Options{
		Host: "127.0.0.1",
		Port: -1, // Puerto aleatorio
	}

	ns, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create NATS server: %v", err)
	}

	go ns.Start()

	if !ns.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server not ready")
	}

	t.Cleanup(func() {
		ns.Shutdown()
		ns.WaitForShutdown()
	})

	return ns, ns.ClientURL()
}

func TestNewClient(t *testing.T) {
	_, url := setupTestNATS(t)

	// Test: crear cliente
	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	// Verificar que el cliente está conectado
	if !client.IsConnected() {
		t.Error("client should be connected")
	}
}

func TestClient_PublishSubscribe(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	subject := "test.subject"
	testData := []byte("test message")
	received := make(chan []byte, 1)

	// Suscribirse
	sub, err := client.Subscribe(subject, func(msg *natslib.Msg) {
		received <- msg.Data
	})
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Dar tiempo a que la suscripción se active
	time.Sleep(100 * time.Millisecond)

	// Publicar
	err = client.Publish(subject, testData)
	if err != nil {
		t.Fatalf("Publish() failed: %v", err)
	}

	// Verificar recepción
	select {
	case data := <-received:
		if string(data) != string(testData) {
			t.Errorf("received %s, want %s", data, testData)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestClient_Request(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	subject := "test.request"
	requestData := []byte("ping")
	responseData := []byte("pong")

	// Configurar responder
	_, err = client.Subscribe(subject, func(msg *natslib.Msg) {
		msg.Respond(responseData)
	})
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// Dar tiempo a que la suscripción se active
	time.Sleep(100 * time.Millisecond)

	// Hacer request con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := client.Request(ctx, subject, requestData)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	if string(response.Data) != string(responseData) {
		t.Errorf("received %s, want %s", response.Data, responseData)
	}
}

func TestClient_RequestTimeout(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	// Request a subject sin responder (debe timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = client.Request(ctx, "no.responder", []byte("test"))
	if err == nil {
		t.Error("Request() should have timed out")
	}
}

func TestClient_Close(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Verificar que está conectado
	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	// Cerrar
	err = client.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verificar que ya no está conectado
	if client.IsConnected() {
		t.Error("client should not be connected after Close()")
	}
}
