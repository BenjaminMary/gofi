INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats,
    description,
    iconName, iconCodePoint, colorName, colorHSL, colorHEX)
VALUES
    (-1, 'Pret', 	'specific', -2, 1, 0, 
        'Utilisable uniquement par le système lors de l''utilisation de la fonction prêt.',
        'lend-hand-coin', 'e921', 'blue grey', '(210,30,40)', '#476685'),
    (-1, 'Emprunt', 	'specific', -1, 1, 0, 
        'Utilisable uniquement par le système lors de l''utilisation de la fonction emprunt.',
        'borrow-hand-coin', 'e922', 'blue grey', '(230,30,40)', '#475285')
;