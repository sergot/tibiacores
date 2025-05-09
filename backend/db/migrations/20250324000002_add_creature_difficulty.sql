-- +goose Up
-- +goose StatementBegin
-- First update any creature names that have changed
UPDATE creatures SET name = 'Nomad (Female)' WHERE name = 'Nomad  (Female)';
UPDATE creatures SET name = 'Nomad (Blue)' WHERE name = 'Nomad  (Blue)';
UPDATE creatures SET name = 'Nomad' WHERE name = 'Nomad  (Basic)';
UPDATE creatures SET name = 'Butterfly (Red)' WHERE name = 'Butterfly  (Red)';
UPDATE creatures SET name = 'Butterfly (Purple)' WHERE name = 'Butterfly  (Purple)';
UPDATE creatures SET name = 'Butterfly (Blue)' WHERE name = 'Butterfly  (Blue)';
UPDATE creatures SET name = 'Horse (Gray)' WHERE name = 'Horse  (Gray)';
UPDATE creatures SET name = 'Horse (Brown)' WHERE name = 'Horse  (Brown)';
UPDATE creatures SET name = 'Horse (Dark Brown)' WHERE name = 'Horse  (Taupe)';

-- Add difficulty column
ALTER TABLE creatures ADD COLUMN difficulty INTEGER;

-- Update difficulties from scraped data
update creatures set difficulty = 0 where name = 'Butterfly (Purple)';
update creatures set difficulty = 0 where name = 'Butterfly (Blue)';
update creatures set difficulty = 0 where name = 'Butterfly (Red)';
update creatures set difficulty = 0 where name = 'Wafer Paper Butterfly';
update creatures set difficulty = 0 where name = 'Dog';
update creatures set difficulty = 0 where name = 'Cat';
update creatures set difficulty = 0 where name = 'Mushroom Sniffer';
update creatures set difficulty = 0 where name = 'Truffle Cook';
update creatures set difficulty = 0 where name = 'Truffle';
update creatures set difficulty = 0 where name = 'Chocolate Blob';
update creatures set difficulty = 0 where name = 'Fruit Drop';
update creatures set difficulty = 0 where name = 'Sugar Cube';
update creatures set difficulty = 0 where name = 'Sugar Cube Worker';
update creatures set difficulty = 0 where name = 'Cream Blob';
update creatures set difficulty = 0 where name = 'Gingerbread Man';
update creatures set difficulty = 0 where name = 'Pigeon';
update creatures set difficulty = 0 where name = 'Northern Pike';
update creatures set difficulty = 0 where name = 'Husky';
update creatures set difficulty = 0 where name = 'Modified Gnarlhound';
update creatures set difficulty = 0 where name = 'Berrypest';

-- Level 1 creatures
update creatures set difficulty = 1 where name IN (
    'Rabbit', 'Chicken', 'Silver Rabbit', 'Agrestic Chicken', 'Sheep', 'Squirrel', 'Deer', 'Pig',
    'Flamingo', 'Parrot', 'Seagull', 'Green Frog', 'Bog Frog', 'Fish', 'Cave Parrot', 'Dromedary',
    'Wisp', 'Penguin', 'Skunk', 'Rat', 'Badger', 'Snake', 'Cave Rat', 'Bat', 'Spider', 'Fox',
    'Wolf', 'Bug', 'Winter Wolf', 'Sandcrawler', 'Troll', 'Island Troll', 'Poison Spider',
    'Frost Troll', 'Wasp', 'Goblin', 'Black Sheep', 'Horse (Brown)', 'Horse (Grey)',
    'Horse (Dark Brown)', 'White Deer', 'Wild Horse', 'Grynch Clan Goblin', 'Undead Jester',
    'White Tiger', 'Honey Elemental'
);

-- Level 2 creatures
update creatures set difficulty = 2 where name IN (
    'Adventurer', 'Redeemed Soul', 'Tainted Soul', 'Hyaena', 'Azure Frog', 'Coral Frog',
    'Crimson Frog', 'Orchid Frog', 'Spit Nettle', 'Bear', 'Panda', 'Swamp Troll', 'Orc',
    'Salamander', 'Polar Bear', 'Crab', 'Cobra', 'Lion', 'Centipede', 'Skeleton',
    'Dworc Venomsniper', 'Emerald Damselfly', 'Crazed Beggar', 'Goblin Scavenger',
    'Orc Spearman', 'Insect Swarm', 'Rotworm', 'Chakoya Tribewarden', 'Tiger',
    'Troll Champion', 'Chakoya Toolshaper', 'Dworc Fleshhunter', 'Crocodile', 'Elf',
    'Larva', 'Scorpion', 'Skeleton Warrior', 'Swampling', 'Dwarf', 'Leaf Golem',
    'Chakoya Windcaller', 'Smuggler', 'Minotaur', 'Marsh Stalker', 'Orc Warrior',
    'Goblin Assassin', 'Dworc Voodoomaster', 'War Wolf', 'Amazon', 'Wild Warrior',
    'Toad', 'Nomad', 'Gnarlhound', 'Boar', 'Minotaur Archer', 'Bandit', 'Poacher',
    'Dwarf Soldier', 'Carrion Worm', 'Slug', 'Gang Member', 'Elf Scout', 'Ghoul',
    'Barbarian Headsplitter', 'Barbarian Skullhunter', 'Valkyrie', 'Pirate Skeleton',
    'Stalker', 'Gazer', 'Barbarian Brutetamer', 'Tortoise', 'Gladiator',
    'Damaged Worker Golem', 'Quara Mantassin Scout', 'Dark Apprentice',
    'Novice of the Cult', 'Assassin', 'Sibang', 'Rorc', 'Orc Shaman', 'Orc Rider',
    'Kongra', 'Ghost', 'Tarnished Spirit', 'Tarantula', 'White Shade', 'Witch',
    'Scarab', 'Pirate Marauder', 'Deepling Worker', 'Dark Monk', 'Fire Devil',
    'Merlkin', 'Hunter', 'Minotaur Mage', 'Mummy', 'Gargoyle', 'Corym Charlatan',
    'Cyclops', 'Frost Giant', 'Frost Giantess', 'Terror Bird', 'Thornback Tortoise',
    'Slime', 'Minotaur Guard', 'Lizard Sentinel', 'Stone Golem', 'Blood Crab',
    'Elephant', 'Mammoth', 'Terramite', 'Dwarf Guard', 'Bonelord', 'Elf Arcanist',
    'Mercury Blob', 'Gozzler', 'Deepsea Blood Crab', 'Furious Troll', 'Dark Magician',
    'Crypt Shambler', 'Monk', 'Abyssal Calamary', 'Mad Scientist', 'Pirate Ghost',
    'Lizard Templar', 'Jellyfish', 'Calamary', 'Damaged Crystal Golem',
    'Little Corym Charlatan', 'Undead Mine Worker', 'Nomad (Female)', 'Nomad (Blue)',
    'Gloom Wolf', 'Undead Prospector', 'Rabid Wolf', 'Filth Toad', 'Mole',
    'Ragged Rabid Wolf', 'Killer Rabbit', 'Terrified Elephant', 'Feverish Citizen',
    'Honour Guard', 'Squidgy Slime', 'Starving Wolf', 'Ghost Wolf', 'Grave Robber',
    'Crypt Defiler', 'Ladybug', 'Firestarter', 'Insectoid Scout', 'Water Buffalo',
    'Troll Guard', 'Goblin Leader', 'Doomsday Cultist', 'Cake Golem', 'Dire Penguin',
    'Acolyte of Darkness', 'Iron Servant'
);

-- Level 3 creatures
update creatures set difficulty = 3 where name IN (
    'Dwarf Henchman', 'Troll Legionnaire', 'Mutated Human', 'Carniphila',
    'Deepling Scout', 'Pirate Cutthroat', 'Dragon Hatchling', 'Orc Berserker',
    'Barbarian Bloodwalker', 'Cyclops Drone', 'Quara Constrictor Scout',
    'Orc Marauder', 'Lizard Snakecharmer', 'Blue Djinn', 'Green Djinn',
    'Fire Elemental', 'Furious Fire Elemental', 'Wilting Leaf Golem',
    'Evil Sheep', 'Demon Skeleton', 'Acid Blob', 'Lancer Beetle',
    'Pirate Buccaneer', 'Cyclops Smith', 'Corym Skirmisher', 'Dwarf Geomancer',
    'Orc Leader', 'Elder Bonelord', 'Zombie', 'Ice Golem', 'Death Blob',
    'Acolyte of the Cult', 'Vampire', 'Haunted Treeling', 'Forest Fury',
    'Swarmer', 'Pirate Corsair', 'Quara Constrictor', 'Adept of the Cult',
    'Clay Guardian', 'Quara Predator Scout', 'Shadow Pupil', 'Efreet',
    'Marid', 'Priestess', 'Mutated Rat', 'Wailing Widow', 'Clomp',
    'Corym Vanguard', 'Pooka', 'Crystalcrusher', 'Enlightened of the Cult',
    'Nightstalker', 'Stonerefiner', 'Wyvern', 'Energy Elemental',
    'Earth Elemental', 'Enraged Crystal Golem', 'Bonebeast', 'Necromancer',
    'Ice Witch', 'Twisted Pooka', 'Quara Pincher Scout', 'Quara Mantassin',
    'Iron Servant Replica', 'Ogre Shaman', 'Dragon Lord Hatchling',
    'Water Elemental', 'Insectoid Worker', 'Orc Warlord', 'Glooth Blob',
    'Pixie', 'Dragon', 'Shark', 'Ancient Scarab', 'Frost Dragon Hatchling',
    'Blood Hand', 'Rot Elemental', 'Mutated Bat', 'Mutated Tiger', 'Omnivora',
    'Stampor', 'Faun', 'Undead Gladiator', 'Roaring Lion', 'Ogre Brute',
    'Quara Hydromancer Scout', 'Vampire Viscount', 'Bog Raider', 'Waspoid',
    'Nymph', 'Cult Believer', 'Orc Cult Minion', 'Blood Priest', 'Lich',
    'Banshee', 'Vicious Squire', 'Dark Faun', 'Wiggler', 'Mooh''Tah Warrior',
    'Crystal Spider', 'Giant Spider', 'Brimstone Bug', 'Killer Caiman',
    'Askarak Demon', 'Putrid Mummy', 'Quara Hydromancer', 'Boogy',
    'Ogre Savage', 'Gravedigger', 'Minotaur Cult Follower', 'Braindeath',
    'Deepling Spellsinger', 'Cult Enforcer', 'Young Sea Serpent',
    'Weakened Frazzlemaw', 'Crawler', 'Blood Beast', 'Vampire Bride',
    'Orc Cult Inquisitor', 'Enfeebled Silencer', 'Orclops Ravager',
    'Parder', 'Massive Water Elemental', 'Massive Earth Elemental',
    'Lizard Legionnaire', 'Spitter', 'Minotaur Cult Prophet',
    'Instable Breach Brood', 'Shaburak Demon', 'Jungle Moa', 'Hero',
    'Lizard Dragon Priest', 'Renegade Knight', 'Drillworm', 'Exotic Bat',
    'Misguided Thief', 'Misguided Bully', 'Iks Aucar', 'Worker Golem',
    'Yielothax', 'Souleater', 'Iks Chuka', 'Nightmare Scion',
    'Minotaur Cult Zealot', 'Instable Sparkion', 'Iks Pututu',
    'Massive Fire Elemental', 'Exotic Cave Spider', 'Orclops Doomhauler',
    'Lizard High Guard', 'Lumbering Carnivor', 'Metal Gargoyle',
    'Worm Priestess', 'High Voltage Elemental', 'Deepling Warrior',
    'Lost Thrower', 'Vile Grandmaster', 'Quara Pincher', 'Wyrm',
    'Werefox', 'Werebadger', 'Pirat Scoundrel', 'Shaper Matriarch',
    'Minotaur Hunter', 'Pirat Bombardier', 'Lizard Zaogun', 'Devourer',
    'Glooth Anemone', 'Lost Husher', 'Lost Exile', 'Broken Shaper',
    'Eternal Guardian', 'Pirat Cutthroat', 'Nightmare', 'Quara Predator',
    'Werewolf', 'Glooth Brigand', 'Stabilizing Dread Intruder',
    'Glooth Golem', 'Stabilizing Reality Reaver', 'Cursed Ape',
    'Wereboar', 'Glooth Bandit', 'Lizard Magistratus', 'Twisted Shaper',
    'Spectre', 'Frost Dragon', 'Dragon Lord', 'Deepling Guard', 'Hydra',
    'Werebear', 'Barkless Devotee', 'Werehyaena Shaman', 'Werehyaena',
    'Sea Serpent', 'Lost Basher', 'Rustheap Golem', 'Pirat Mate',
    'Execowtioner', 'Draken Warmaster', 'Barkless Fanatic', 'Destroyer',
    'Behemoth', 'Lizard Chosen', 'Hellspawn', 'Moohtant', 'War Golem',
    'Enslaved Dwarf', 'Diabolic Imp', 'Serpent Spawn', 'Betrayed Wraith',
    'Midnight Asura', 'Plaguesmith', 'Warlock', 'Ravenous Lava Lurker',
    'Lost Soul', 'Frost Flower Asura', 'Feversleep', 'Silencer',
    'Vampire Pig', 'Hot Dog', 'Infernal Frog', 'Doom Deer', 'Flying Book',
    'Berserker Chicken', 'Demon Parrot', 'Deepling Brawler', 'Arctic Faun',
    'Evil Sheep Lord', 'Noble Lion', 'Elder Forest Fury', 'Swan Maiden',
    'Hibernal Moth', 'Lacewing Moth', 'Massive Energy Elemental',
    'Orc Cultist', 'Orc Cult Priest', 'Gryphon', 'Orc Cult Fanatic',
    'Cult Scholar', 'Stone Rhino', 'Goldhanded Cultist',
    'Goldhanded Cultist Bride', 'Dragonling', 'Walker', 'Deepling Elite',
    'Iks Churrascan', 'Manta Ray', 'Ghoulish Hyaena', 'Tomb Servant',
    'Roast Pork', 'Sacred Spider', 'Animated Snowman', 'Cow',
    'Baleful Bunny', 'Grave Guard', 'Bellicose Orger', 'Orger',
    'Elder Mummy', 'Loricate Orger', 'Schiach', 'Percht',
    'Sandstone Scorpion', 'Death Priest', 'Renegade Quara Mantassin',
    'Askarak Lord', 'Renegade Quara Constrictor', 'Golden Servant Replica',
    'Diamond Servant Replica', 'Askarak Prince', 'Minotaur Invader',
    'Renegade Quara Hydromancer', 'Shaburak Lord',
    'Deepling Master Librarian', 'Renegade Quara Pincher', 'Ice Dragon',
    'Shaburak Prince', 'Kollos', 'Spidris', 'Renegade Quara Predator',
    'Spidris Elite', 'Dryad', 'Thornfire Wolf', 'Nightslayer',
    'Crystal Wolf', 'Elf Overseer', 'Bane Bringer', 'Bride of Night',
    'Herald of Gloom', 'Golden Servant', 'Yeti', 'Undead Cavebear',
    'Shadow Hound', 'Diamond Servant', 'Midnight Warrior', 'Bane of Light',
    'Midnight Panther', 'Midnight Spawn', 'Vicious Manbat', 'Raging Fire',
    'Iks Ahpututu', 'Crustacea Gigantica', 'Nightfiend', 'Albino Dragon',
    'Draptor', 'Duskbringer'
);

-- Level 4 creatures
update creatures set difficulty = 4 where name IN (
    'Vulcongra', 'Sparkion', 'Spiky Carnivor', 'Breach Brood',
    'Lizard Noble', 'Menacing Carnivor', 'Minotaur Amazon', 'Werelion',
    'White Lion', 'Werelioness', 'Shock Head', 'Iks Yapunac',
    'Dread Intruder', 'Reality Reaver', 'Elder Wyrm', 'Deepworm',
    'Goggle Cake', 'Nibblemaw', 'Diremaw', 'Humongous Fungus',
    'Candy Horror', 'Angry Sugar Fairy', 'Draken Spellweaver',
    'Mitmah Scout', 'Two-Headed Turtle', 'Cave Devourer',
    'Ripper Spectre', 'Foam Stalker', 'Werepanther', 'Chasm Spawn',
    'Dawnfire Asura', 'Cunning Werepanther', 'Defiler',
    'Hideous Fungus', 'Frazzlemaw', 'Hellfire Fighter',
    'Candy Floss Elemental', 'Weretiger', 'Infernalist', 'Fury',
    'Lava Lurker', 'Medusa', 'Retching Horror', 'Werecrocodile',
    'Bulltaur Alchemist', 'Gazer Spectre', 'Ogre Rowdy',
    'Magma Crawler', 'Phantasm', 'Dark Carnisylvan',
    'Poisonous Carnisylvan', 'Tunnel Tyrant', 'Flimsy Lost Soul',
    'Draken Abomination', 'Juvenile Bashmu', 'Mitmah Seer',
    'Ghastly Dragon', 'Dark Torturer', 'Arachnophobica',
    'Crazed Winter Rearguard', 'Crazed Summer Rearguard',
    'Choking Fear', 'Bulltaur Brute', 'Hulking Carnisylvan',
    'Draken Elite', 'Manticore', 'Lost Berserker',
    'Skeleton Elite Warrior', 'Crazed Summer Vanguard',
    'Ogre Ruffian', 'Hand of Cursed Fate', 'Bashmu', 'Crape Man',
    'Dragolisk', 'Undead Elite Gladiator', 'Naga Archer',
    'White Weretiger', 'Cursed Prospector', 'Young Goanna',
    'Venerable Girtablilu', 'Blemished Spawn',
    'Crazed Winter Vanguard', 'Feral Werecrocodile', 'Hellhound',
    'Grim Reaper', 'Ogre Sage', 'Mean Lost Soul', 'Rhindeer',
    'Makara', 'Harpy', 'Soul-broken Harbinger', 'Girtablilu Warrior',
    'Wardragon', 'Naga Warrior', 'Son of Verminor', 'Insane Siren',
    'Burster Spectre', 'Demon', 'Eyeless Devourer', 'Guzzlemaw',
    'Crypt Warrior', 'Adult Goanna', 'Demon Outcast', 'Lavafungus',
    'Vexclaw', 'Streaked Devourer', 'Deathling Scout', 'Thanatursus',
    'Falcon Knight', 'Varnished Diremaw', 'Liodile',
    'Bulltaur Forgepriest', 'Deathling Spellsinger', 'Blightwalker',
    'Priestess of the Wild Sun', 'Lavaworm', 'Afflicted Strider',
    'Carnivostrich', 'Sineater Inferniarch', 'Mega Dragon',
    'Usurper Archer', 'Cave Chimera', 'Usurper Knight',
    'Falcon Paladin', 'Tremendous Tyrant', 'Cobra Assassin',
    'Sphinx', 'Usurper Warlock', 'Freakish Lost Soul',
    'True Frost Flower Asura', 'Boar Man', 'Gorger Inferniarch',
    'Black Sphinx Acolyte', 'Grimeleech', 'Terrorsleep',
    'Cobra Scout', 'True Midnight Asura', 'Burning Gladiator',
    'Broodrider Inferniarch', 'True Dawnfire Asura', 'Undead Dragon',
    'Cobra Vizier', 'Floating Savant', 'Hellhunter Inferniarch',
    'Quara Raider', 'Spellreaper Inferniarch', 'Crypt Warden',
    'Feral Sphinx', 'Quara Looter', 'Evil Prospector', 'Lamassu',
    'Rootthing Nutshell', 'Biting Book', 'Rootthing Bug Tracker',
    'Animated Feather', 'Quara Plunderer', 'Juggernaut',
    'Energuardian of Tales', 'Hellflayer', 'Energetic Book',
    'Rootthing Amber Shaper', 'Icecold Book', 'Burning Book',
    'Ink Blob', 'Squid Warden', 'Rage Squid', 'Sight of Surrender',
    'Brain Squid', 'Brinebrute Inferniarch', 'Infected Weeper',
    'Armadile', 'Stone Devourer', 'Ironblight', 'Weeper',
    'Orewalker', 'Cliff Strider', 'Lava Golem', 'Guardian of Tales',
    'Knowledge Elemental', 'Cursed Book', 'Seacrest Serpent',
    'Deepling Tyrant', 'Hive Overseer', 'Haunted Dragon'
);

-- Level 5 creatures
update creatures set difficulty = 5 where name IN (
    'Sulphur Spouter', 'Stalking Stalk', 'Mantosaurus', 'Sabretooth',
    'Headpecker', 'Mercurial Menace', 'Emerald Tortoise', 'Gore Horn',
    'Nighthunter', 'Hulking Prehemoth', 'Gorerilla', 'Noxious Ripptor',
    'Sulphider', 'Undertaker', 'Shrieking Cry-Stal', 'Brachiodemon',
    'Infernal Phantom', 'Turbulent Elemental', 'Capricious Phantom',
    'Branchy Crawler', 'Rotten Golem', 'Bony Sea Devil',
    'Infernal Demon', 'Mould Phantom', 'Druid''s Apparition',
    'Knight''s Apparition', 'Paladin''s Apparition',
    'Sorcerer''s Apparition', 'Monk''s Apparition',
    'Distorted Phantom', 'Many Faces', 'Courage Leech',
    'Vibrant Phantom', 'Cloak of Terror', 'Darklight Emitter',
    'Oozing Corpus', 'Oozing Carcass', 'Mycobiontic Beetle',
    'Converter', 'Bloated Man-Maggot', 'Meandering Mushroom',
    'Darklight Construct', 'Darklight Striker', 'Darklight Matter',
    'Darklight Source', 'Sopping Corpus', 'Rotten Man-Maggot',
    'Wandering Pillar', 'Sopping Carcass', 'Walking Pillar'
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE creatures DROP COLUMN difficulty;
-- +goose StatementEnd 