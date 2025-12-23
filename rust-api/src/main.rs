use actix_web::{post, web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize)]
struct Venta {
    categoria: String,
    producto_id: String,
    precio: f64,
    cantidad_vendida: i32,
}

async fn enviar_a_go(venta: &Venta) -> Result<(), reqwest::Error> {
    let client = reqwest::Client::new();

    client
        .post("http://localhost:8081/procesar")
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


#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .service(recibir_venta) 
    })
    .bind(("0.0.0.0", 8080))?
    .workers(4)
    .run()
    .await
}
