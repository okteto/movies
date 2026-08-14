const express = require("express");
const mongo = require("mongodb").MongoClient;
const { requireAdmin } = require("./session");

const DEFAULT_COPIES = 3;

const app = express();
app.use(express.json());

const url = `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}?authSource=admin`;

const toMovie = (body) => {
  const copies = Number.parseInt(body.copies, 10);
  return {
    original_title: String(body.original_title ?? "").trim(),
    overview: String(body.overview ?? "").trim(),
    backdrop_path: String(body.backdrop_path ?? "").trim(),
    price: Number.parseFloat(body.price) || 0,
    vote_average: Number.parseFloat(body.vote_average) || 0,
    copies: Number.isInteger(copies) && copies >= 0 ? copies : DEFAULT_COPIES,
  };
};

const invalid = (movie) => {
  if (!movie.original_title) {
    return "the movie needs a title";
  }
  if (movie.price < 0) {
    return "the price can't be negative";
  }
  return null;
};

async function startWithRetry() {
  try {
    const client = await mongo.connect(url, { 
      connectTimeoutMS: 30000,
      socketTimeoutMS: 30000,
    });

    const db = client.db(process.env.MONGODB_DATABASE);
    const catalog = db.collection('catalog');

    app.get("/catalog/healthz", (req, res, next) => {
      res.sendStatus(200)
      return;
    });

    app.get("/catalog", async (req, res, next) => {
      console.log(`GET /catalog`)
      try {
        const results = await catalog.find().sort({ id: 1 }).toArray();
        res.json(results.map((movie) => ({ copies: DEFAULT_COPIES, ...movie })));
      } catch (err) {
        console.log(`failed to query movies: ${err}`)
        res.json([]);
      }
    });

    app.post("/catalog", requireAdmin, async (req, res, next) => {
      const movie = toMovie(req.body ?? {});
      const error = invalid(movie);
      if (error) {
        res.status(400).json({ error });
        return;
      }

      try {
        const [last] = await catalog.find().sort({ id: -1 }).limit(1).toArray();
        const id = (last?.id ?? 0) + 1;
        await catalog.insertOne({ _id: id, id, ...movie });
        console.log(`POST /catalog created movie ${id}`);
        res.status(201).json({ id, ...movie });
      } catch (err) {
        console.log(`failed to create movie: ${err}`);
        res.status(500).json({ error: "failed to create the movie" });
      }
    });

    app.put("/catalog/:id", requireAdmin, async (req, res, next) => {
      const id = Number.parseInt(req.params.id, 10);
      const movie = toMovie(req.body ?? {});
      const error = invalid(movie);
      if (error) {
        res.status(400).json({ error });
        return;
      }

      try {
        const result = await catalog.updateOne({ id }, { $set: movie });
        if (result.matchedCount === 0) {
          res.status(404).json({ error: "movie not found" });
          return;
        }
        console.log(`PUT /catalog/${id}`);
        res.json({ id, ...movie });
      } catch (err) {
        console.log(`failed to update movie ${id}: ${err}`);
        res.status(500).json({ error: "failed to update the movie" });
      }
    });

    app.delete("/catalog/:id", requireAdmin, async (req, res, next) => {
      const id = Number.parseInt(req.params.id, 10);

      try {
        const result = await catalog.deleteOne({ id });
        if (result.deletedCount === 0) {
          res.status(404).json({ error: "movie not found" });
          return;
        }
        console.log(`DELETE /catalog/${id}`);
        res.json({ id });
      } catch (err) {
        console.log(`failed to delete movie ${id}: ${err}`);
        res.status(500).json({ error: "failed to delete the movie" });
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
