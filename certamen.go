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
	colaprocesos []bcp
	contadorDisp int
}

func (d *dispatcher) agregarProceso(pid int, estado string, contadorProg int, nombreproceso string) {
	d.colaprocesos = append(d.colaprocesos, bcp{pid, estado, contadorProg, nombreproceso})
}

func (d *dispatcher) ejecutarDispatcher(procesoActual chan bcp, tiempoEjecucion int) {

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

	var wgReaders sync.WaitGroup
	wgReaders.Add(2)

	go func() {
		defer wgReaders.Done()
		for numero := range numeros {
			Tiempo_ejecucion = append(Tiempo_ejecucion, numero)
		}
	}()

	go func() {
		defer wgReaders.Done()
		for linea := range lineas {
			fmt.Println(linea)
		}
	}()

	wg.Wait()
	wgReaders.Wait()
	indice := 0
	x := Tiempo_ejecucion[indice]
	var d dispatcher

	for i := 1; i <= 10; i++ {
		fmt.Println("iteracion: ", i)
		if i == x {
			d.agregarProceso(i, "Listo", 0, "Proceso"+strconv.Itoa(i))
			fmt.Println("Proceso", i, "agregado a la cola de procesos")
			x = Tiempo_ejecucion[indice+1]
		}
	}

	fmt.Println("Números guardados en el slice:", Tiempo_ejecucion)

	fmt.Println("Contenido de dispatcher:")
	for _, proceso := range d.colaprocesos {
		fmt.Printf("PID: %d, Estado: %s, ContadorProg: %d, NombreProceso: %s\n",
			proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)
	}

}
