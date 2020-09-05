 //This code was imagined, directed,developed and commented in spanish by Jose Santos L.
package main;

import(
	"fmt"
	"net"
	"net/http"
	"html/template"
	"os"
	"io/ioutil"
	"log"
	"bufio"
	"io"
	"strings"
	"github.com/mdp/qrterminal"
)

var Nota []string;//Este string llevara los apuntes necesarios dentro de cada 

type PageData struct{//Estructura de los datos para mostrar en el index.
	Bloc []string//Aqui se mete el string la variable publica Nota.
	Archivos []string //Aqui va la lista de nombres de archivos.
}

//FUNCIONES  TOMADAS DE https://stackoverflow.com/questions/5884154/read-text-file-into-string-array-and-write 
//Funcion para leer un archivo como []string
func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}
//Funcion para escribir un archivo como []string
func writeLines(lines []string, path string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
    return w.Flush()
}

/*En principio la funcion readNota solo debe ser usado una vez, al inicio del
programa. Este se encarga de guardar los apuntes de nota.txt en la variable
publica Nota.*/
func readNota() {
	if _, err := os.Stat("nota.txt"); err == nil {//Comprueba si el archivo existe
		Nota,_ =readLines("nota.txt");
	} else {//si el archivo no existe.
		f,_ := os.Create("nota.txt");//se crea uno nuevo
		f.Close()//se cierra por las mismas este archivo
		fmt.Println("El archivo de notas no existe.se ha creado uno nuevo");
	}
	fmt.Println(Nota);
}

//Esta funcion lee todos los archivos almacenados para el template
func listArchivos() []string{
	var archivos []string;
	files,_:=ioutil.ReadDir("./archivos");

	for _,f := range files{
		archivos = append(archivos,f.Name())
	}
	return archivos;
}

//Retorna la direccion ip en forma de string.
func getIp() string{
	conn, err := net.Dial("udp", "8.8.8.8:80")
	var ip string;
	if err != nil{//muestra un posible error
		log.Println(err);
		ip="127.0.0.1:8080";
	}else{
		defer conn.Close();//cierra conexion.
		localAddr := conn.LocalAddr().(*net.UDPAddr);//consigue la direccion local.
		ip = localAddr.String();//convierte la direccion en texto.
		sep:=strings.Index(ip,":");//encuentra el separador.
		ip=ip[:sep];//quita el puerto anterior.
		ip="http://"+ip+":8080";//se pone bonito con puerto y todo.
	}
	return ip;
}

//DE AQUI PARA ABAJO SON HANDLERS DE DIRECCION
//simplemente ordena todos los datos y envia el template del index.
func index_handler(w http.ResponseWriter, r *http.Request){
	tmpl, _ := template.ParseFiles("index.html");
	archivos:=listArchivos();
	data := PageData{Bloc:Nota,Archivos:archivos};
	tmpl.Execute(w,data);
}

func writeNota(w http.ResponseWriter, r *http.Request){
	r.ParseForm();//Obtener datos del formulario
	Nota = r.Form["bloc"];//Meter el mensaje en la nota
	fmt.Println(Nota);//Imprime la nota.
	_ = writeLines(Nota,"nota.txt");//COnserva la nota en un archivo.
	http.Redirect(w,r,"/",301);
}

//FUncion que se encarga de subir los archivos cargados a la carpeta.
func upload(w http.ResponseWriter, r *http.Request){
	r.ParseForm();
	file, header, err := r.FormFile("archivo");
	switch err {
	case nil:
		log.Println(header.Filename +" fue cargado.");
		break
	case http.ErrMissingFile:
		log.Println("Ningun archivo fue cargado.");
		http.Redirect(w,r,"/",301);
		break
	default:
		log.Fatal(err);
		http.Redirect(w,r,"/",301);
	}
	defer file.Close();//eventualmente se cerrara.
	//Se guarda el archivo en la carpeta de archivos.
	f, err := os.OpenFile("./archivos/" + header.Filename, os.O_WRONLY | os.O_CREATE, 0666);
	if err != nil{
		log.Fatal(err);
	}
	defer f.Close();
	io.Copy(f,file);
	http.Redirect(w,r,"/",301);//Redirije a la pagina de inicio.
}
func del_handler(w http.ResponseWriter, r *http.Request){
	arc:=r.URL.Path[5:];//se limpia la url del nombre del archivo
	if arc != ""{
		arc="./archivos/"+arc;//se estaablece la ubicacion del archivo.
		if _,err := os.Stat(arc); err == nil{
			err:=os.Remove(arc);
			if err == nil{
				log.Println(arc," fue eliminado");
			}
			http.Redirect(w,r,"/",301);
				
		}else{
			fmt.Fprintf(w,"file not found.");
		}
			
	}else{fmt.Fprintf(w,"Ningun archivo fue pasado.");}
}
func download_handler(w http.ResponseWriter, r *http.Request){
	arc:=r.URL.Path[3:];//se limpia la url del nombre del archivo
	if arc != ""{
		arc="./archivos/"+arc;//se estaablece la ubicacion del archivo.
		if _,err := os.Stat(arc); err == nil{
			http.ServeFile(w,r,arc);//se envia el archivo..
			http.Redirect(w,r,"/",301);
		}else{
			fmt.Fprintf(w,"file not found.");
		}
			
	}else{fmt.Fprintf(w,"Ningun archivo fue pasado.");}
}

func main() {
	fmt.Println("Iniciando sistema Transfer.");
	ip:=getIp();
	qrterminal.Generate(ip, qrterminal.L, os.Stdout)
	fmt.Println(ip);
	readNota();//Pasa los apuntes de nota.txt a la variable Nota.
	http.HandleFunc("/",index_handler);//lleva a la pantalla de incio.
	http.HandleFunc("/paste",writeNota);
	http.HandleFunc("/upload",upload);
	http.HandleFunc("/del/",del_handler);//funcion de eliminar archivo.
	http.HandleFunc("/d/",download_handler);//handler para descargar archivos
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080",nil);
}
