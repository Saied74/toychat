I have been reading Alex Edwards book "Let's Go" and building this application.
I have used it in the past to build other applications, but here, I am working
to build something bigger and also learn more of the ideas in the book.
I have also had my eye on doing something with the NATS server.  So, I am
combining some of the things that I have been looking to learn into this project.

As of now, this application is composed of five components:
Web: which is the front end facing the end user.
Chat: is a simple simulation that reverses the order of words in a sentence
Mat: is both a simple and stupid application that replaces every word with mat
dbmgr: is the database interface for the user database
nats server: which connects the web, chat, mat and dbmgr together.
mysql database: which provides a table for the session manager and a table for
end user information.

Right now, the dbmgr, chat and mat are not threaded.  Next step is to thread
them with go routines.  
