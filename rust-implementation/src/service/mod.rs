use chrono::NaiveDateTime;
use sqlx::{Pool, Postgres};

use crate::models::WeatherData;

#[derive(thiserror::Error, Debug)]
pub enum WeatherServiceError {
    #[error("Database Error Occurred")]
    DatabaseError(#[from] sqlx::Error),
}

#[derive(Clone)]
pub struct WeatherService(pub Pool<Postgres>);

impl WeatherService {
    pub async fn fetch_range(
        &self,
        start: NaiveDateTime,
        end: NaiveDateTime,
    ) -> Result<Vec<WeatherData>, WeatherServiceError> {
        Ok(sqlx::query_as!(
            WeatherData,
            "SELECT * FROM measurements WHERE time >= $1 AND time <= $2 ORDER BY time ASC",
            start,
            end
        )
        .fetch_all(&self.0)
        .await?)
    }

    pub async fn create_measurement(
        &self,
        measurement: WeatherData,
    ) -> Result<(), WeatherServiceError> {
        sqlx::query!(
            "INSERT INTO measurements (time, temperature, humidity) VALUES ($1, $2, $3) ON CONFLICT (time) DO UPDATE SET humidity=excluded.humidity, temperature=excluded.temperature",
            measurement.time,
            measurement.temperature,
            measurement.humidity
        )
        .execute(&self.0)
        .await?;
        Ok(())
    }
}
