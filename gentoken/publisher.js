require("dotenv").config();
const jwt = require("jsonwebtoken");

const secret = process.env.MERCURE_JWT_PUBLISHER;
const topic = process.env.MERCURE_TOPIC;

const payload = {
    mercure: {
        publish: [topic],
    },
    // exp: Math.floor(Date.now() / 1000) + 60 * 60, // expired 1 jam
};
console.log(payload, secret, topic)
const token = jwt.sign(payload, secret, { algorithm: "HS256" });

console.log("Publisher JWT:\n", token);
