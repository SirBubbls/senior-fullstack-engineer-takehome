use axum::{
    Extension, Router,
    http::StatusCode,
    routing::{any, get, post},
};
use models::WeatherData;
use service::WeatherService;
use sqlx::postgres::PgPoolOptions;
use std::{env, sync::Arc};
use tokio::sync::broadcast::{Receiver, Sender};

mod handlers;
pub mod models;
pub mod service;

#[derive(Clone)]
struct AppContext {
    service: WeatherService,
    updates: Arc<Receiver<WeatherData>>,
    update_sender: Sender<WeatherData>,
}

#[tokio::main]
async fn main() {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");

    let db = PgPoolOptions::new()
        .max_connections(5)
        .connect(&database_url)
        .await
        .unwrap();
    let service = WeatherService(db);

    let (tx, rx) = tokio::sync::broadcast::channel::<models::WeatherData>(50);
    let context = AppContext {
        service,
        updates: Arc::new(rx),
        update_sender: tx,
    };

    let app = Router::new()
        .route("/health", get(|| async { StatusCode::OK }))
        .route("/day", get(handlers::date_single))
        .route("/range", get(handlers::date_range))
        .route("/submit", post(handlers::submit_data))
        .route("/updates", any(handlers::update_ws))
        .layer(Extension(context));

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
