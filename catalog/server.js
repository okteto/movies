const express = require("express");
const mongo = require("mongodb").MongoClient;

const app = express();

const url = `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}?authSource=admin`;

async function startWithRetry() {
  try {
    const client = await mongo.connect(url, { 
      connectTimeoutMS: 30000,
      socketTimeoutMS: 30000,
    });

    const db = client.db(process.env.MONGODB_DATABASE);

    app.get("/catalog/healthz", (req, res, next) => {
      res.sendStatus(200)
      return;
    });

    app.get("/catalog", async (req, res, next) => {
      console.log(`GET /catalog`)
      try {
        const results = await db.collection('catalog').find().toArray();
        res.json(results);
      } catch (err) {
        console.log(`failed to query movies: ${err}`)
        res.json([]);
      }
    });

    app.listen(8080, () => {
      console.log("Server running on port 8080.");
    });
  } catch (err) {
    console.error(`Error connecting, retrying in 1 sec: ${err}`);
    setTimeout(startWithRetry, 1000);
  }
};

startWithRetry();