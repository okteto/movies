const mongo = require("mongodb").MongoClient;

const url = `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}`;

var insert = function(collection, data, resolve, reject) {
  const d = require(data);
  d.results.forEach((doc) => {
    doc._id = doc.id;
  });
  collection.insertMany(d.results, (err, r) => {
    if (err) {
      if (err.code != 11000) {
        return reject(err);
      }
    }

    resolve();
  });
}

mongo.connect(url, { 
  useUnifiedTopology: true,
  useNewUrlParser: true,
  reconnectTries: 30,
  reconnectInterval: 1000
}, (err, client) => {
  if (err) {
    console.error(`Error connecting to mongodb, retrying in 1 sec: ${err}`);
    process.exit(1);
  }

  var promises = [];
  db = client.db(process.env.MONGODB_DATABASE);
  promises.push(new Promise((resolve, reject)=>{
    insert(db.collection('movies'), "./data/movies.json", resolve, reject);
  }));

  promises.push(new Promise((resolve, reject)=>{
    insert(db.collection('watching'), "./data/watching.json", resolve, reject);
  }));

  Promise.all(promises)
  .then(function() { 
    console.log('all loaded'); 
    process.exit(0);
  })
  .catch((err) => {
    console.error(`fail to load: ${err}`);
    process.exit(1);
  });
});