const { MongoClient } = require("mongodb");

const url = process.env.MONGODB_USERNAME && process.env.MONGODB_PASSWORD 
  ? `mongodb://${process.env.MONGODB_USERNAME}:${encodeURIComponent(process.env.MONGODB_PASSWORD)}@${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}?authSource=admin`
  : `mongodb://${process.env.MONGODB_HOST}:27017/${process.env.MONGODB_DATABASE}`;

async function insertData(collection, dataPath) {
  const data = require(dataPath);
  data.results.forEach((doc) => {
    doc._id = doc.id;
  });
  
  try {
    await collection.insertMany(data.results);
    console.log(`Inserted ${data.results.length} documents`);
  } catch (err) {
    if (err.code !== 11000) {
      throw err;
    }
    console.log('Documents already exist, skipping insertion');
  }
}

async function loadWithRetry() {
  const client = new MongoClient(url, {
    connectTimeoutMS: 30000,
    socketTimeoutMS: 30000,
    serverSelectionTimeoutMS: 30000,
  });

  try {
    console.log('Connecting to MongoDB...');
    await client.connect();
    console.log('Connected successfully');

    const db = client.db(process.env.MONGODB_DATABASE);
    await insertData(db.collection('catalog'), "./data/catalog.json");
    
    console.log('All data loaded successfully');
    await client.close();
    process.exit(0);
  } catch (err) {
    console.error(`Error: ${err}`);
    await client.close();
    
    console.log('Retrying in 3 seconds...');
    setTimeout(loadWithRetry, 3000);
  }
}

loadWithRetry();
