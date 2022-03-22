const express = require("express");
const mongo = require("mongodb").MongoClient;

const app = express();

const url = `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:3940/${process.env.MONGODB_DATABASE}`;

function startWithRetry() {
  mongo.connect(url, { 
    useUnifiedTopology: true,
    useNewUrlParser: true,
    connectTimeoutMS: 1000,
    socketTimeoutMS: 1000,
  }, (err, client) => {
    if (err) {
      console.error(`Error connecting, retrying in 1 sec: ${err}`);
      setTimeout(startWithRetry, 1000);
      return;
    }

    const db = client.db(process.env.MONGODB_DATABASE);

    app.listen(8080, () => {
      app.get("/api/healthz", (req, res, next) => {
        res.sendStatus(200)
        return;
      });

      app.get("/api/movies", (req, res, next) => {
        console.log(`GET /api/movies`)
        db.collection('movies').find().toArray( (err, results) =>{
          if (err){
            console.log(`failed to query movies: ${err}`)
            res.json([]);
            return;
          }
          res.json(results);
        });
      });

      app.get("/api/watching", (req, res, next) => {
        console.log(`GET /api/watching`)
        db.collection('movies').find().toArray( (err, results) =>{
          if (err){
            console.log(`failed to query watching: ${err}`)
            res.json([]);
            return;
          }

          res.json(results);
        });
      });

      console.log("Server running on port 8080.");
    });
  });
};

startWithRetry();