# jwtAuth
Api имеет 2 REST пути:
1. /auth/Tokens - обращение через POST, в теле HTTP передавать json с полем GUID, в ответ записывает куки с access и refresh токенами
2. /auth/Refresh - обращение через PUT, с HTTP передавать куки с access и refresh токенами, в ответ обновляет оба токена


