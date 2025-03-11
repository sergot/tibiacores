// Connect to the fiendlist database
db = db.getSiblingDB('fiendlist');

// Read the creatures.json file
const fs = require('fs');
const creaturesData = JSON.parse(fs.readFileSync('/docker-entrypoint-initdb.d/creatures.json', 'utf8'));

// Insert creatures into the database
if (creaturesData && creaturesData.creatures && Array.isArray(creaturesData.creatures)) {
  console.log(`Importing ${creaturesData.creatures.length} creatures...`);
  
  // Create a collection for creatures if it doesn't exist
  if (!db.getCollectionNames().includes('creatures')) {
    db.createCollection('creatures');
  }
  
  // Insert each creature with upsert to avoid duplicates
  creaturesData.creatures.forEach(creature => {
    db.creatures.updateOne(
      { endpoint: creature.endpoint },
      { $set: creature },
      { upsert: true }
    );
  });
  
  console.log('Creatures imported successfully!');
} else {
  console.log('No creatures found in the JSON file or invalid format.');
} 