package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type bcp struct {
	pid           int
	estado        string
	contadorProg  int
	nombreproceso string
}

type dispatcher struct {
	colaprocesos   []bcp
	colabloqueados []bcp
	contadorDisp   int
}

func (d *dispatcher) agregarProceso(pid int, estado string, contadorProg int, nombreproceso string) {
	d.colaprocesos = append(d.colaprocesos, bcp{pid, estado, contadorProg, nombreproceso})
}

func (d *dispatcher) ejecutarDispatcher(procesoActual chan bcp, tiempoEjecucion int) {

}

func (d *dispatcher) bloquearProceso() {
	procesoBloqueado := d.colaprocesos[0]
	d.colabloqueados = append(d.colabloqueados, procesoBloqueado)
	d.colaprocesos = d.colaprocesos[1:]
	fmt.Printf("Proceso %d bloqueado\n", procesoBloqueado.pid)
}

func (d *dispatcher) desbloquearProceso() {
	procesoDesbloqueado := d.colabloqueados[0]
	d.colaprocesos = append(d.colaprocesos, procesoDesbloqueado)
	d.colabloqueados = d.colabloqueados[1:]
}

func (d *dispatcher) finalizarProceso(procesoFinalizado chan<- bool) {
	go func() {
		d.colaprocesos[0].estado = "Finalizado"
		d.colaprocesos = d.colaprocesos[1:]
		procesoFinalizado <- true
	}()
}

func (d *dispatcher) ejecutarProceso(procesoActual <-chan *bcp, tiempoEjecucion int, done chan<- bool, procesoFinalizado chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for proceso := range procesoActual {
		nombreProceso := proceso.nombreproceso

		file, err := os.Open(nombreProceso)
		if err != nil {
			fmt.Println("Error al abrir el archivo:", err)
			done <- false
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		linea := 0
		lineasEjecutadas := 0
		encontrado := false
		var tiempoBloqueo int

		for scanner.Scan() {
			linea++
			if linea == proceso.contadorProg {
				encontrado = true
			}
			if encontrado {

				if d.contadorDisp == tiempoBloqueo {
					d.desbloquearProceso()
					tiempoBloqueo = 0

				}
				fmt.Println("iteracion: ", d.contadorDisp)
				err := AgregarLinea("Salida.txt", strconv.Itoa(d.contadorDisp)+"					 "+scanner.Text()+"					"+proceso.nombreproceso+"				"+strconv.Itoa(proceso.contadorProg))
				if err != nil {
					fmt.Println("Error al agregar línea:", err)
				}
				d.contadorDisp++
				proceso.contadorProg++
				if len(d.colaprocesos) > 1 {
					d.colaprocesos[1].contadorProg = d.colaprocesos[1].contadorProg + 1
				} else {
					d.colaprocesos[0].contadorProg = d.colaprocesos[0].contadorProg + 0
				}

				fmt.Println(scanner.Text())
				if strings.HasPrefix(scanner.Text(), "F") {
					if len(d.colaprocesos) > 1 {
						d.colaprocesos[1].estado = "Finalizado"
						done <- true
						break
					} else {
						d.colaprocesos[0].estado = "Finalizado"
						done <- true
						break
					}

				}

				lineasEjecutadas++
				if lineasEjecutadas == tiempoEjecucion {
					fmt.Printf("Contador actualizado del proceso %d: %d\n", proceso.pid, proceso.contadorProg)
					done <- true
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error al leer el archivo:", err)
			done <- false
		}
	}
}

func leerArchivo(nombreArchivo string, lineas chan<- string, numeros chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(nombreArchivo)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		close(lineas)
		close(numeros)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linea := scanner.Text()
		lineas <- linea

		words := strings.Fields(linea)
		if len(words) > 0 {
			numero, err := strconv.Atoi(words[0])
			if err == nil {
				numeros <- numero
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo:", err)
	}

	close(lineas)
	close(numeros)
}

// funcion para agregar lineas al archivo
func AgregarLinea(rutaArchivo, linea string) error {
	archivo, err := os.OpenFile(rutaArchivo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer archivo.Close()

	escritor := bufio.NewWriter(archivo)

	_, err = escritor.WriteString(linea + "\n")
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}
	err = escritor.Flush()
	if err != nil {
		return fmt.Errorf("error al flush del búfer: %v", err)
	}

	return nil
}

func main() {
	lineas := make(chan string)
	numeros := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)

	go leerArchivo("Orden_creacion.txt", lineas, numeros, &wg)
	var Tiempo_ejecucion []int
	var nombre_proceso []string

	var wgReaders sync.WaitGroup
	wgReaders.Add(2)

	go func() {
		defer wgReaders.Done()
		for numero := range numeros {
			Tiempo_ejecucion = append(Tiempo_ejecucion, numero)
		}
	}()
	// Guardar nombre de los procesos en el arreglo
	go func() {
		defer wgReaders.Done()
		for linea := range lineas {
			fmt.Println(linea)
			words := strings.Fields(linea)
			nombre_proceso = append(nombre_proceso, words[1])
		}
	}()

	wg.Wait()
	wgReaders.Wait()
	err := AgregarLinea("Salida.txt", "#Tiempo de CPU		Tipo Instrucción		Proceso/Despachador		Valor CP")
	if err != nil {
		fmt.Println("Error al agregar línea:", err)
	}

	indice := 0
	var d dispatcher
	d.contadorDisp = 1
	procesoActual := make(chan *bcp, len(Tiempo_ejecucion))
	done := make(chan bool)
	procesoFinalizado := make(chan bool)
	wg.Add(1)

	go d.ejecutarProceso(procesoActual, 5, done, procesoFinalizado, &wg)

	for i := 1; i < 6; i++ {

		// Agregar procesos a la cola
		if indice < len(Tiempo_ejecucion) {
			x := Tiempo_ejecucion[indice]
			if d.contadorDisp == x || d.contadorDisp < x {
				proceso := bcp{
					pid:           100 + indice,
					estado:        "Listo",
					contadorProg:  1,
					nombreproceso: nombre_proceso[indice],
				}

				d.agregarProceso(proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)

				indice++
			}
		}
		if len(d.colaprocesos) > 1 {
			err := AgregarLinea("Salida.txt", strconv.Itoa(d.contadorDisp)+"						PULL					Dispacher ")
			d.contadorDisp++
			err2 := AgregarLinea("Salida.txt", strconv.Itoa(d.contadorDisp)+"						LOAD					 "+d.colaprocesos[0].nombreproceso+"		")
			d.contadorDisp++
			err3 := AgregarLinea("Salida.txt", strconv.Itoa(d.contadorDisp)+"						EXEC					Dispacher ")
			d.contadorDisp++
			if err != nil || err2 != nil || err3 != nil {
				fmt.Println("Error al agregar línea:", err)
			}
			proceso := &d.colaprocesos[0]
			d.colaprocesos = append(d.colaprocesos[1:], *proceso)
			procesoActual <- proceso
			<-done
		}
		select {
		case <-procesoFinalizado:
			fmt.Println("Un proceso ha finalizado")
			if len(d.colaprocesos) == 1 {
				for {
					wg.Add(1)
					proceso := &d.colaprocesos[0]
					d.colaprocesos = append(d.colaprocesos[1:], *proceso)
					procesoActual <- proceso
				}
			} else {
				fmt.Println("No hay más procesos en la cola")
			}
		default:
		}

	}

	// Cerrar canal y esperar gorrutina
	close(procesoActual)
	wg.Wait()
	close(done)
	close(procesoFinalizado)

	// Información final del dispatcher
	fmt.Println("Números guardados en el slice:", Tiempo_ejecucion)
	fmt.Println("Nombres guardados en el slice:", nombre_proceso)
	fmt.Println("Contenido de dispatcher:")
	for _, proceso := range d.colaprocesos {
		fmt.Printf("PID: %d, Estado: %s, ContadorProg: %d, NombreProceso: %s\n",
			proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)
	}
	fmt.Println("Contador dispatcher: ", d.contadorDisp)
}
