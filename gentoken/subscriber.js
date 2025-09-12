require("dotenv").config();
const jwt = require("jsonwebtoken");

const secret = process.env.MERCURE_JWT_SUBSCRIBER;
const topic = process.env.MERCURE_TOPIC;

const payload = {
    mercure: {
        subscribe: [topic],
    },
    // exp: Math.floor(Date.now() / 1000) + 60 * 60, // expired 1 jam
};

const token = jwt.sign(payload, secret, { algorithm: "HS256" });

console.log("Subscriber JWT:\n", token);
