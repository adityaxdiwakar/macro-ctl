#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;

use serde::Serialize;
use rocket::response::content;
use std::process::Command;


#[derive(Serialize)]
struct Response {
    code: i16,
    message: String
}

#[get("/instruct/off")]
fn off() -> content::Json<std::string::String> {
    Command::new("shutdown")
        .spawn()
        .expect("failed to execute process");

    
    let resp = Response{
        code: 200,
        message: "Powering off within 60 seconds".to_string()
    };
    content::Json(serde_json::to_string(&resp).unwrap())
}

#[get("/instruct/pulse")]
fn pulse() -> content::Json<std::string::String> {
    Command::new("python3")
        .arg("pulse.py")
        .spawn()
        .expect("failed to execute process");


    let resp = Response{
        code: 200,
        message: "Rebooting pulseaudio, standby!".to_string()
    };
    content::Json(serde_json::to_string(&resp).unwrap())
}

fn main() {
    rocket::ignite().mount("/", routes![off, pulse]).launch();
}
