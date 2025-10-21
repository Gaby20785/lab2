package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"sync"
	"net"
	"time"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "lab2/nodos/proto"
)

// Nodo - Estructura que representa un nodo en el sistema
type NodoDB struct {
	pb.UnimplementedCyberDayServiceServer
	nombre        	string
	direccion       string
	ofertas       	[]*pb.OfertaRequest
	mu            	sync.Mutex
	contadorOfertas int
	probFallo 		float64
	enFallo       	bool
	caidasSimuladas int
	startTime     	time.Time
	client          pb.CyberDayServiceClient
}

// registrarEnBroker - Registra el nodo actual en el broker.
func (n *NodoDB) registrarEnBroker(nodoID string) {

	resp, err := n.client.RegistrarNodo(context.Background(), &pb.RegistroNodoRequest{
		Nombre: nodoID,
		Direccion: n.direccion,
	})
	if err != nil {
		log.Printf("Error registrando %s: %v", nodoID, err)
		return
	}

	if resp.GetExito() {
		log.Printf("%s registrado en broker", nodoID)
	} else {
		log.Printf("Registro de %s falló", nodoID)
	}
}

// EnviarOferta - Recibe y almacena una oferta en el nodo.
// Si el nodo está en fallo, la oferta es rechazada. También simula fallos aleatorios.
// el primer nodo puede fallar en los primeros 30 segundos desde que se reciben ofertas, 
// el segundo nodo puede fallar desde los 40 segundos hasta los 70 segundos y el último nodo
// puede fallar despues de los 80 segundos de ejecución hasta la finalización de la ejecución.
func (n *NodoDB) EnviarOferta(ctx context.Context, req *pb.OfertaRequest) (*pb.OfertaResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.startTime.IsZero() {
		n.startTime = time.Now()
		log.Printf("%s - Timer iniciado con primera oferta", n.nombre)
	}

	// Si está en fallo, no procesar
	if n.enFallo {
		log.Printf("%s en fallo - rechazando oferta", n.nombre)
		return &pb.OfertaResponse{Exito: false}, nil
	}

	elapsed := time.Since(n.startTime)
	puedeFallar := false

	switch n.nombre {
	case "DB1":
		puedeFallar = elapsed <= 30*time.Second
	case "DB2":
		puedeFallar = elapsed >= 40*time.Second && elapsed <= 70*time.Second
	case "DB3":
		puedeFallar = elapsed >= 80*time.Second
	}

	if !n.enFallo && n.probFallo > 0 && puedeFallar {
		if rand.Float64() < n.probFallo {
			n.simularFallo()
			return &pb.OfertaResponse{Exito: false}, nil
		}
	}

	for _, ofertaExistente := range n.ofertas {
		if ofertaExistente.GetOfertaId() == req.GetOfertaId() {
			log.Printf("%s: Oferta duplicada %s - ignorando", n.nombre, req.GetOfertaId())
			return &pb.OfertaResponse{Exito: true}, nil
		}
	}

	n.contadorOfertas++
	n.ofertas = append(n.ofertas, req)

	log.Printf("%s almacenó: %s - $%d", n.nombre, req.GetProducto(), req.GetPrecio())
	log.Printf("   - Total en %s: %d ofertas", n.nombre, len(n.ofertas))

	return &pb.OfertaResponse{Exito: true}, nil
}

// simularFallo - Marca el nodo como caído e inicia un proceso de recuperación automática.
func (n *NodoDB) simularFallo() {
	n.enFallo = true
	n.caidasSimuladas++
	
	log.Printf("%s CAÍDA SIMULADA - Probabilidad: %.1f%%", 
		n.nombre, n.probFallo*100)
	log.Printf("   - Caída #%d - Recuperación en 5 segundos", n.caidasSimuladas)

	go n.recuperarAutomaticamente()
}

// recuperarAutomaticamente - Espera 5 segundos e intenta resincronizar el nodo.
func (n *NodoDB) recuperarAutomaticamente() {
	log.Printf("%s programado para recuperarse en 5 segundos", n.nombre)
	time.Sleep(5 * time.Second)

	log.Printf("%s iniciando resincronización...", n.nombre)
	exito := n.solicitarResincronizacion()
	
	n.mu.Lock()
    if exito {
        n.enFallo = false
        log.Printf("%s COMPLETAMENTE RECUPERADO Y SINCRONIZADO", n.nombre)
    } else {
        log.Printf("%s falló resincronización - reintentando en 5s", n.nombre)
        go func() {
            time.Sleep(5 * time.Second)
            n.recuperarAutomaticamente()
        }()
    }
    n.mu.Unlock()
}

// solicitarResincronizacion - Solicita al broker las ofertas faltantes para ponerse al día.
// Devuelve true si la sincronización fue exitosa.
func (n *NodoDB) solicitarResincronizacion() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

    n.mu.Lock()
    ofertasActuales := n.ofertas
    n.mu.Unlock()

    resp, err := n.client.SincronizarEntidad(ctx, &pb.SincronizacionRequest{
        EntidadId:      n.nombre,
        Tipo:           "nodo",
        OfertasActuales: ofertasActuales,
    })
    
    if err != nil || !resp.GetExito() {
        log.Printf("%s error en resincronización: %v", n.nombre, err)
        return false
    }

    n.mu.Lock()
    ofertasRecibidas := 0
    for _, oferta := range resp.GetOfertasFaltantes() {
        existe := false
        for _, ofertaExistente := range n.ofertas {
            if ofertaExistente.GetOfertaId() == oferta.GetOfertaId() {
                existe = true
                break
            }
        }
        if !existe {
            n.ofertas = append(n.ofertas, oferta)
            ofertasRecibidas++
        }
    }
    n.contadorOfertas = len(n.ofertas)
    n.mu.Unlock()

    log.Printf("%s resincronizado: +%d ofertas", n.nombre, ofertasRecibidas)
    return true
}

// LeerOfertas - Devuelve todas las ofertas almacenadas en el nodo.
// Si el nodo está caído, responde con exito: false.
func (n *NodoDB) LeerOfertas(ctx context.Context, req *pb.LecturaRequest) (*pb.LecturaResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.enFallo {
		return &pb.LecturaResponse{Exito: false}, nil
	}

	log.Printf("%s enviando %d ofertas", n.nombre, len(n.ofertas))
	
	return &pb.LecturaResponse{
		Ofertas: n.ofertas,
		Exito:   true,
	}, nil
}

// main - Configura el nodo.
func main() {
	var nodoID string
	var direccion string
	flag.StringVar(&nodoID, "nodo", "", "ID del nodo DB (DB1, DB2, DB3)")
	flag.Parse()

	if nodoID == "" {
		log.Fatal("Debe especificar el nodo: --nodo=DB1|DB2|DB3")
	}

	direccion = os.Getenv("NODO_DIRECCION")

	puerto := ""
	var probFallo float64

	switch nodoID {
	case "DB1":
		puerto = ":50052"
		probFallo = 0.1
	case "DB2":
		puerto = ":50053"
		probFallo = 0.1
	case "DB3":
		puerto = ":50054"
		probFallo = 0.1
	default:
		log.Fatalf("Nodo no válido: %s", nodoID)
	}

	if direccion == "" {
		switch nodoID {
		case "DB1":
			direccion = "localhost:50052"
		case "DB2":
			direccion = "localhost:50053"
		case "DB3":
			direccion = "localhost:50054"
		}
	}

	rand.Seed(time.Now().UnixNano())

	log.Printf("Iniciando nodo: %s en %s", nodoID, direccion)
	log.Printf("Probabilidad de fallo: %.1f%%", probFallo*100)

	brokerHost := os.Getenv("BROKER_HOST")
	if brokerHost == "" {
		brokerHost = "broker"  // nombre del servicio en docker-compose
	}
	conn, err := grpc.Dial(brokerHost + ":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	
	if err != nil {
		log.Fatalf("No se pudo conectar al broker: %v", err)
	}
	defer conn.Close()

	client := pb.NewCyberDayServiceClient(conn)

	nodo := &NodoDB{
		nombre:           nodoID,
		direccion:        direccion,
		ofertas:          make([]*pb.OfertaRequest, 0),
		contadorOfertas:  0,
		probFallo: probFallo,
		enFallo:          false,
		caidasSimuladas:  0,
		startTime:        time.Time{},
		client:           client,
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterCyberDayServiceServer(grpcServer, nodo)

	listenAddress := "0.0.0.0" + puerto
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("Error al iniciar nodo %s: %v", nodoID, err)
	}
	
	go nodo.registrarEnBroker(nodoID)

	log.Printf("Nodo %s listo en puerto %s", nodoID, puerto)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error en servidor nodo %s: %v", nodoID, err)
	}

}
