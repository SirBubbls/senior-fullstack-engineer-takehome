use axum::{
    Extension, Json,
    extract::{Query, WebSocketUpgrade},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use chrono::ParseError;
use serde::Deserialize;

use crate::{
    AppContext,
    models::{CreateWeatherData, WeatherData, parse_custom_time},
    service::WeatherServiceError,
};

#[derive(thiserror::Error, Debug)]
pub enum RequestError {
    #[error("Resource Not Found")]
    NotFound,
    #[error("Malformed Date Format")]
    MalformedData(#[from] ParseError),
    #[error("Invalid Data")]
    InvalidData,
    #[error("Internal Service Error")]
    InternalService(#[from] WeatherServiceError),
}

impl IntoResponse for RequestError {
    fn into_response(self) -> Response {
        match self {
            RequestError::NotFound => StatusCode::NOT_FOUND.into_response(),
            RequestError::MalformedData(e) => {
                        (StatusCode::BAD_REQUEST, e.to_string()).into_response()
                    }
            RequestError::InternalService(_) => StatusCode::INTERNAL_SERVER_ERROR.into_response(),
            RequestError::InvalidData => StatusCode::BAD_REQUEST.into_response(),
        }
    }
}

#[derive(Deserialize)]
pub struct DayParams {
    day: String,
}
pub async fn date_single(
    Query(params): Query<DayParams>,
    Extension(AppContext { service, .. }): Extension<AppContext>,
) -> Result<Json<WeatherData>, RequestError> {
    let start = parse_custom_time(params.day)?;
    if let Some(day) = service.fetch_range(start, start).await?.pop() {
        return Ok(Json(day));
    }
    Err(RequestError::NotFound)
}

#[derive(Deserialize)]
pub struct RangeParams {
    start: String,
    end: String,
}
pub async fn date_range(
    Query(params): Query<RangeParams>,
    Extension(AppContext { service, .. }): Extension<AppContext>,
) -> Result<Json<Vec<WeatherData>>, RequestError> {
    let start = parse_custom_time(params.start)?;
    let end = parse_custom_time(params.end)?;
    Ok(Json(service.fetch_range(start, end).await?))
}

pub async fn submit_data(
    Extension(AppContext {
        service,
        update_sender,
        ..
    }): Extension<AppContext>,
    Json(data): Json<CreateWeatherData>,
) -> Result<StatusCode, RequestError> {
    if !data.is_valid() {
        return Err(RequestError::InvalidData)
    }
    let data: WeatherData = data.try_into()?;
    let _ = update_sender.send(data.clone());
    service.create_measurement(data).await?;
    Ok(StatusCode::OK)
}

pub async fn update_ws(
    Extension(AppContext { updates, .. }): Extension<AppContext>,
    ws: WebSocketUpgrade,
) -> Response {
    let mut receiver = updates.resubscribe();
    ws.on_upgrade(move |mut socket| async move {
        while let Ok(update) = receiver.recv().await {
            socket
                .send(axum::extract::ws::Message::Text(
                    serde_json::to_string(&update).unwrap().into(),
                ))
                .await
                .unwrap();
        }
    })
}
