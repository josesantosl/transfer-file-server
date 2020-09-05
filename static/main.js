console.log("javascript conectado");
validate = function(){ //funcion ejecutada cuando se aplasta el submit de archivo
    if( document.getElementById("archivo").files.length == 0 ){
	alert('Ningun archivo fue cargado');
    }else{
	document.getElementById('fileform').action = "/upload"; 
    }
    document.getElementById('fileform').submit()
}
