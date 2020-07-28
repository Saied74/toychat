//Copyright (c) 2020 Saied Seghatoleslami
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
The front end web application is a fairly simple web server following the
guidlelines of Alex Edewards Let's Go.  The only part needing in my opinion is
the playHander.  matHandler and playMatHandler are just test cases that I will
eventurlaly delete.

playHander is an Ajax handler getting message from the JavaScript code in
the browser transmitted on click and waits for a reply.  This is how it is
handled:
1. extract the user id from the request context.
2. The middlware has already varified that it exists.
3. Use broker.GetDialog to get the Dialog object.
4. Check to see if one exists and if an agent is allocated
5. If dialog does not exist, use broker.MakeDialog to make one.
6. If agent does not assigned or it is a new dialog ask for agent.
7. This is done by calling broker.SelectAgent.
8. After the agent is selected, the dialog table by calling broker.UpdateDialog
9.// TODO: max agent, agent sill and that stuff needs work in the future.
10. Once dialog exists and agent is assigned, store message using broker.StoreMsg
11. Send message to "Agent-"+AgentID
11. Messages are sent via broker.SendMsg, wait for reply.
12. Send reply back through the Ajax interface.

*/

package main
