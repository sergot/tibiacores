-- +goose Up
-- +goose StatementBegin

-- Add 48 new creature(s)
INSERT INTO creatures (name) VALUES ('Bluebeak') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Bramble Wyrmling') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Cinder Wyrmling') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Corrupted Ghost') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Corrupted Skeleton') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Creepy Crawler') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Crusader') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Crypt Construct') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Crypt Fiend') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Crypt Mage') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Cyclursus') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Dworc Shadowstalker') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Gloom Maw') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Haunted Hunter') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Hawk Hopper') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Headwalker') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Imperial') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Infernoid Blob') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Infernoid Hound') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Infernoid Soul') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Infernoid Spiritual') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Ink Splash') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lion Hydra') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lizard Commander') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lizard Executioner') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lizard Henchman') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lizard Magician') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Lizard Swordmaster') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Monk (Creature)') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Muglex Clan Assassin') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Muglex Clan Footman') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Night Harpy') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Norcferatu Heartless') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Norcferatu Nightweaver') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Orclops Bloodbreaker') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Pirate Cook') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Pirate Gunner') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Pirate Navigator') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Pirate Quartermaster') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Raubritter Chastener') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Raubritter Marksman') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Raubritter Skirmisher') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Roaming Dread') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Sea Captain') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Shell Drake') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Stag') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Varg') ON CONFLICT (name) DO NOTHING;
INSERT INTO creatures (name) VALUES ('Walking Dread') ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM creatures WHERE name = 'Bluebeak';
DELETE FROM creatures WHERE name = 'Bramble Wyrmling';
DELETE FROM creatures WHERE name = 'Cinder Wyrmling';
DELETE FROM creatures WHERE name = 'Corrupted Ghost';
DELETE FROM creatures WHERE name = 'Corrupted Skeleton';
DELETE FROM creatures WHERE name = 'Creepy Crawler';
DELETE FROM creatures WHERE name = 'Crusader';
DELETE FROM creatures WHERE name = 'Crypt Construct';
DELETE FROM creatures WHERE name = 'Crypt Fiend';
DELETE FROM creatures WHERE name = 'Crypt Mage';
DELETE FROM creatures WHERE name = 'Cyclursus';
DELETE FROM creatures WHERE name = 'Dworc Shadowstalker';
DELETE FROM creatures WHERE name = 'Gloom Maw';
DELETE FROM creatures WHERE name = 'Haunted Hunter';
DELETE FROM creatures WHERE name = 'Hawk Hopper';
DELETE FROM creatures WHERE name = 'Headwalker';
DELETE FROM creatures WHERE name = 'Imperial';
DELETE FROM creatures WHERE name = 'Infernoid Blob';
DELETE FROM creatures WHERE name = 'Infernoid Hound';
DELETE FROM creatures WHERE name = 'Infernoid Soul';
DELETE FROM creatures WHERE name = 'Infernoid Spiritual';
DELETE FROM creatures WHERE name = 'Ink Splash';
DELETE FROM creatures WHERE name = 'Lion Hydra';
DELETE FROM creatures WHERE name = 'Lizard Commander';
DELETE FROM creatures WHERE name = 'Lizard Executioner';
DELETE FROM creatures WHERE name = 'Lizard Henchman';
DELETE FROM creatures WHERE name = 'Lizard Magician';
DELETE FROM creatures WHERE name = 'Lizard Swordmaster';
DELETE FROM creatures WHERE name = 'Monk (Creature)';
DELETE FROM creatures WHERE name = 'Muglex Clan Assassin';
DELETE FROM creatures WHERE name = 'Muglex Clan Footman';
DELETE FROM creatures WHERE name = 'Night Harpy';
DELETE FROM creatures WHERE name = 'Norcferatu Heartless';
DELETE FROM creatures WHERE name = 'Norcferatu Nightweaver';
DELETE FROM creatures WHERE name = 'Orclops Bloodbreaker';
DELETE FROM creatures WHERE name = 'Pirate Cook';
DELETE FROM creatures WHERE name = 'Pirate Gunner';
DELETE FROM creatures WHERE name = 'Pirate Navigator';
DELETE FROM creatures WHERE name = 'Pirate Quartermaster';
DELETE FROM creatures WHERE name = 'Raubritter Chastener';
DELETE FROM creatures WHERE name = 'Raubritter Marksman';
DELETE FROM creatures WHERE name = 'Raubritter Skirmisher';
DELETE FROM creatures WHERE name = 'Roaming Dread';
DELETE FROM creatures WHERE name = 'Sea Captain';
DELETE FROM creatures WHERE name = 'Shell Drake';
DELETE FROM creatures WHERE name = 'Stag';
DELETE FROM creatures WHERE name = 'Varg';
DELETE FROM creatures WHERE name = 'Walking Dread';

-- +goose StatementEnd