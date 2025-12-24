use actix_web::{post, web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
use std::io::Write;

#[derive(Deserialize, Serialize)]
struct Venta {
    categoria: String,
    producto_id: String,
    precio: f64,
    cantidad_vendida: i32,
}

async fn enviar_a_go(venta: &Venta) -> Result<(), reqwest::Error> {
    let go_service_url = std::env::var("GO_SERVICE_URL")
        .unwrap_or_else(|_| "http://go-api:8081".to_string());
    
    let client = reqwest::Client::new();

    client
        .post(&format!("{}/procesar", go_service_url))
        .json(venta)
        .send()
        .await?;

    Ok(())
}

#[post("/venta")]
async fn recibir_venta(venta: web::Json<Venta>) -> impl Responder {
    if let Err(e) = enviar_a_go(&venta).await {
        eprintln!("Error enviando a Go: {}", e);
        return HttpResponse::InternalServerError().body("Error enviando a Go");
    }

    HttpResponse::Ok().json("Venta procesada")
}

#[actix_web::main] // Solo una vez
async fn main() -> std::io::Result<()> {
    println!("Iniciando Servidor en puerto 8080...");
    std::io::stdout().flush().unwrap(); 

    HttpServer::new(|| {
        App::new()
            .service(recibir_venta)
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}