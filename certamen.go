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

func (d *dispatcher) bloquearProceso(tiempoEnbloqueo int) {
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

func (d *dispatcher) ejecutarProceso(procesoActual <-chan *bcp, tiempoEjecucion int, done chan<- bool, wg *sync.WaitGroup) {
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

				if strings.HasPrefix(scanner.Text(), "E/S") {
					words := strings.Fields(scanner.Text())
					if len(words) > 1 {
						var err error
						tiempoBloqueo, err = strconv.Atoi(words[1])
						if err == nil {
							fmt.Printf("Valor de E/S: %d\n", tiempoBloqueo)
							d.bloquearProceso(tiempoBloqueo)
							break
						} else {
							fmt.Println("Error al convertir el número:", err)
						}
					}
				}

				if d.contadorDisp == tiempoBloqueo {
					d.desbloquearProceso()
					tiempoBloqueo = 0

				}

				fmt.Println(scanner.Text())
				lineasEjecutadas++
				if lineasEjecutadas == tiempoEjecucion {
					proceso.contadorProg = linea + 1
					fmt.Printf("Contador actualizado del proceso %d: %d\n", proceso.pid, proceso.contadorProg)
					d.contadorDisp = linea + 1
					fmt.Printf("Contador del dispatcher actualizado: %d\n", d.contadorDisp)
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
	indice := 0
	var d dispatcher
	d.contadorDisp = 1
	procesoActual := make(chan *bcp, len(Tiempo_ejecucion))
	done := make(chan bool)
	wg.Add(1)

	go d.ejecutarProceso(procesoActual, 5, done, &wg)

	for i := 1; i <= 6; i++ {
		fmt.Println("iteracion: ", d.contadorDisp)

		// Agregar procesos a la cola
		if indice < len(Tiempo_ejecucion) {
			x := Tiempo_ejecucion[indice]
			if d.contadorDisp == x {
				proceso := bcp{
					pid:           100 + indice,
					estado:        "Listo",
					contadorProg:  1,
					nombreproceso: nombre_proceso[indice],
				}
				d.agregarProceso(proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)
				fmt.Println("Proceso", proceso.nombreproceso, "agregado a la cola de procesos")
				indice++
			}
		}

		// Procesar procesos en la cola
		if len(d.colaprocesos) > 0 {
			// Enviar el puntero del primer proceso en la cola a la goroutine
			proceso := &d.colaprocesos[0]
			procesoActual <- proceso

			<-done

			// Mover el proceso al final de la cola
			d.colaprocesos = append(d.colaprocesos[1:], *proceso)
		}
	}

	// Cerrar canal y esperar gorrutina
	close(procesoActual)
	wg.Wait()
	close(done)

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
