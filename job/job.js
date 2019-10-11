const mongo = require("mongodb").MongoClient;

const url = `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}`;

function startWithRetry() {
  mongo.connect(url, { 
    useUnifiedTopology: true,
    useNewUrlParser: true,
    reconnectTries: 30,
    reconnectInterval: 1000
  }, (err, client) => {
    if (err) {
      console.error(`Error connecting to mongodb, retrying in 1 sec: ${err}`);
      setTimeout(startWithRetry, 1000);
      return;
    }

    const db = client.db(process.env.MONGODB_DATABASE);
    const data = {
      movies: require("./data/movie.json"),
      mylist: require("./data/mylist.json"),
      watching: require("./data/watching.json")
    };
  
    createNewEntries(db.collection('movies'), data.movies.results);
    createNewEntries(db.collection('mylist'), data.mylist.results);
    createNewEntries(db.collection('watching'), data.watching.results);
    process.exit(0);
    });
};

var createNewEntries = function(collection, entries) {
  entries.forEach(function(doc) {
    doc._id = doc.id;
    collection.insertOne(doc, (err, res) => {
      if (err) {
        if (err.code != 11000) {
          console.log(`error while writing to mongo: ${err}`);
        } else {
          var query = { _id: doc.id };
          var values = { $set: doc};
          collection.updateOne(query, values, (err, res) => {
            if (err) {
              console.log(`error while updating to mongo: ${err}`);
            }
          });
        }
      } 
    });
  });
};

startWithRetry();