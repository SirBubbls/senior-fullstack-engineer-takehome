use chrono::ParseError;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize)]
pub struct WeatherData {
    pub humidity: f32,
    pub temperature: f32,
    pub time: chrono::NaiveDateTime,
}

#[derive(Debug, Clone, Deserialize)]
pub struct CreateWeatherData {
    pub humidity: f32,
    pub temperature: f32,
    pub date: String,
}

impl TryFrom<CreateWeatherData> for WeatherData {
    type Error = chrono::ParseError;

    fn try_from(value: CreateWeatherData) -> Result<Self, Self::Error> {
        Ok(WeatherData {
            humidity: value.humidity,
            temperature: value.temperature,
            time: parse_custom_time(value.date)?,
        })
    }
}

impl CreateWeatherData {
    pub fn is_valid(&self) -> bool {
        if self.humidity < 0.0 || self.humidity > 100.0 {
            return false;
        }
        true
    }
}

pub fn parse_custom_time(time: String) -> Result<chrono::NaiveDateTime, ParseError> {
    Ok(chrono::NaiveDate::parse_from_str(&time, "%Y-%m-%d")?.into())
}

#[cfg(test)]
mod tests {
    use chrono::Datelike as _;

    use super::parse_custom_time;

    #[test]
    fn parse_date() {
        let input = "2023-01-02";
        let parsed = parse_custom_time(input.to_string()).unwrap();
        assert_eq!(parsed.day(), 2);
        assert_eq!(parsed.month(), 1);
        assert_eq!(parsed.year(), 2023);
    }
}
