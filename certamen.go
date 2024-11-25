package main

import (
	"bufio"
	"fmt"
	"os"
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

func leerArchivo(ruta string, lineas chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(ruta)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		close(lineas)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineas <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo:", err)
	}

	close(lineas)
}

func main() {
	lineas := make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go leerArchivo("Orden_creacion.txt", lineas, &wg)

	go func() {
		for linea := range lineas {
			fmt.Println(linea)
		}
	}()

	wg.Wait()
}
