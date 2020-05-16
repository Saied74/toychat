Project objective:  Build a robust and reliable messaging application that can
connect people to either automations or to people who can help them.  The roles
in the system are:

Super admin who can add admins to the system, activate or deactivate them.
The super admin role is added by the offline application su through a CLI.
At this time, deleting the admin must be done using mysql interface
directly.  To change the super admin password, for now, the super admin needs
to be deleted using the MySQL direct interface and added back using the su application.
At this time, I am contemplating only one super user but that can change.

Admin who can add agents, active or deactivate them and change his or her own
password.  There can be multiple admins.   Admins can take action only if they
are in the active state which is controlled by the super user.

Agent who can go online, go offline, and engage in a chat - maybe I will add
telephony in the future - or change their own password only if they are in
the active state..  There can be multiple agents.  Agents are added, made
active or inactive by the admins.

End users who can chat with the system.

Automations that can do useful work - whatever that might be.

Work in progress: current state of the project.

Nats messaging is install and working.  
MySQL with the admins table is up and working.
The front end is the first thing that I built, but it is just test functionality
now.
The backend is partially built.
  Super admin role is complete
  Admin role still needs change password
  Agent role only has login and logout working

In addition to completing the roles, there is a lot of boilerplate code that
needs to be refactored, the packages need to be rationalized, and all code
has to be reviewed for at least following the SOLID principals.


Older notes:
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
