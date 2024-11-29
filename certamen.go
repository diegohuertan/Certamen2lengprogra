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
	tiempoBloqueo int
}

type dispatcher struct {
	colaprocesos   []bcp
	colabloqueados []bcp
	contadorDisp   int
}

func (d *dispatcher) agregarProceso(pid int, estado string, contadorProg int, nombreproceso string) {
	d.colaprocesos = append(d.colaprocesos, bcp{pid, estado, contadorProg, nombreproceso, 0})
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
				for i := 1; i < len(words); i++ {
					numeros <- numero
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo:", err)
	}

	close(lineas)
	close(numeros)
}

// funcion para leer linea
func LeerLinea(nombreArchivo string, numeroLinea int) (string, error) {

	archivo, err := os.Open(nombreArchivo)
	if err != nil {
		return "", fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer archivo.Close()

	scanner := bufio.NewScanner(archivo)

	lineaActual := 1
	for scanner.Scan() {
		if lineaActual == numeroLinea {
			return scanner.Text(), nil
		}
		lineaActual++
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error al leer el archivo: %v", err)
	}
	return "", fmt.Errorf("línea %d no encontrada", numeroLinea)
}
func procesarDispatcher(d *dispatcher, Tiempo_ejecucion []int, nombre_proceso []string, wg *sync.WaitGroup) {
	defer wg.Done()
	indice := 0
	for i := 1; i <= 5; i++ {
		if indice < len(Tiempo_ejecucion) {
			for z := 0; z < len(Tiempo_ejecucion); z++ {
				x := Tiempo_ejecucion[z]
				if d.contadorDisp == x {
					d.agregarProceso(100+indice, "Listo", 0, nombre_proceso[indice])
					fmt.Println("Proceso", nombre_proceso[indice], "agregado a la cola de procesos")
					indice++
				}
			}
		}
		if len(d.colaprocesos) > 0 {
			linea, err := LeerLinea(d.colaprocesos[0].nombreproceso, d.colaprocesos[0].contadorProg+1)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			d.colaprocesos[0].contadorProg++
			fmt.Println("iteracion: ", d.contadorDisp, linea)
			if linea == "F" {
				d.colaprocesos = d.colaprocesos[1:]
			}

			if strings.HasPrefix(linea, "E/S") {
				words := strings.Fields(linea)
				if len(words) > 1 {
					numero, err := strconv.Atoi(words[1])
					if err == nil {
						// Bloquear el proceso usando el número
						d.colabloqueados = append(d.colabloqueados, d.colaprocesos[0])
						d.colaprocesos = d.colaprocesos[1:]
						d.colabloqueados[0].tiempoBloqueo = numero
					} else {
						fmt.Println("Error al convertir el número:", err)
					}
				}
			}

			if len(d.colabloqueados) > 0 {
				d.colabloqueados[0].tiempoBloqueo--
				if d.colabloqueados[0].tiempoBloqueo == 0 {
					d.colaprocesos = append([]bcp{d.colabloqueados[0]}, d.colaprocesos...)
					d.colabloqueados = d.colabloqueados[1:]
					fmt.Println("orden cola", d.colaprocesos)
				}
			}

			if len(d.colaprocesos) == 0 {
				fmt.Println("cortando")
				break
			}
		}

		if i == 5 {
			if len(d.colaprocesos) > 1 {
				primero := d.colaprocesos[0]
				d.colaprocesos = append(d.colaprocesos[1:], primero)
			}
			i = 0
		}

		d.contadorDisp++
	}
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
	go func() {
		defer wgReaders.Done()
		for linea := range lineas {
			words := strings.Fields(linea)
			for i := 1; i < len(words); i++ {
				nombre_proceso = append(nombre_proceso, words[i])
			}
		}
	}()

	wg.Wait()
	wgReaders.Wait()

	var d dispatcher
	wg.Add(1)
	go procesarDispatcher(&d, Tiempo_ejecucion, nombre_proceso, &wg)

	wg.Wait()

	fmt.Println("\nNúmeros guardados en el slice:", Tiempo_ejecucion)
	fmt.Println("nombre guardados en el slice:", nombre_proceso)

	// Informacion del dispatcher
	fmt.Println("Contenido de dispatcher:")
	for _, proceso := range d.colaprocesos {
		fmt.Printf("PID: %d, Estado: %s, ContadorProg: %d, NombreProceso: %s\n",
			proceso.pid, proceso.estado, proceso.contadorProg, proceso.nombreproceso)
		fmt.Println("Contador dispatcher: ", d.contadorDisp)
	}
}
