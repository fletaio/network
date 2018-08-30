package main

import (
	"fleta"
	"fleta/mock/mocknetwork"
)

func main() {
	var fleta fleta.Fleta

	mocknetwork.Fleta(&fleta)
	mocknetwork.Run()

	// test6()
}

// func test6() {
// 	addr1 := samutil.Sha256HexInt(0) + ":3000"
// 	addr2 := samutil.Sha256HexInt(1) + ":3000"
// 	addr3 := samutil.Sha256HexInt(2) + ":3000"
// 	addr4 := samutil.Sha256HexInt(3) + ":3000"

// 	consumer := make(chan message.Message)
// 	flanetConsumer := make(chan message.Message)

// 	//handler chaining
// 	h1 := flanetwork.NewMessageHandler(nil, flanetConsumer)
// 	h2 := discovery.NewMessageHandler(h1, consumer)

// 	//init with start handler
// 	pm := peer.NewManager(h2)

// 	var fleta fleta.Fleta
// 	mocknetwork.Fleta(&fleta)

// 	l, err := mocknet.Listen("tcp", addr1)
// 	if err != nil {
// 		panic(err)
// 	}
// 	go func() {
// 		for {
// 			conn1, _ := l.Accept()
// 			//add seed peer connection
// 			pm.AddPeer(peer.NodeID(conn1.RemoteAddr().String()+"_1"), conn1)
// 		}
// 	}()

// 	conn2, _ := mocknet.Dial("tcp", addr1, addr2)
// 	pm.AddPeer(peer.NodeID(addr2), conn2)
// 	conn3, _ := mocknet.Dial("tcp", addr1, addr3)
// 	pm.AddPeer(peer.NodeID(addr3), conn3)
// 	conn4, _ := mocknet.Dial("tcp", addr1, addr4)
// 	pm.AddPeer(peer.NodeID(addr4), conn4)

// 	go func() {
// 		for {
// 			msg := <-consumer
// 			log.Println("consumer msg : ", msg)
// 		}
// 	}()
// 	go func() {
// 		for {
// 			msg := <-flanetConsumer
// 			go func(msg message.Message) {
// 				mType, _ := flanetwork.TypeOfMessage(msg)
// 				switch mType {
// 				case flanetwork.AskFormulatorMessageType:
// 					log.Println("AskFormulatorMessageType msg : ", msg)
// 					af := flanetwork.NewAnswerFormulator(addr1, flanetinterface.FormulatorNode)
// 					payload := message.ToPayload(flanetwork.AnswerFormulatorMessageType, af)
// 					addr := msg.(*flanetwork.AskFormulator).Addr

// 					// f.DialTo(addr)
// 					pm.Send(peer.NodeID(addr), &payload, packet.COMPRESSED)
// 				case flanetwork.AnswerFormulatorMessageType:
// 					log.Println("AnswerFormulatorMessageType msg : ", msg)
// 				}
// 			}(msg)

// 		}
// 	}()

// 	askf := flanetwork.NewAskFormulator(addr1)
// 	payload := message.ToPayload(flanetwork.AskFormulatorMessageType, askf)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr2), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr3), &payload, packet.COMPRESSED)
// 	pm.Send(peer.NodeID(addr4), &payload, packet.COMPRESSED)

// 	for {
// 		time.Sleep(time.Hour * 10)
// 	}
// }

// func test5() {

// 	//test consumer!
// 	consumer := make(chan message.Message)

// 	//handler chaining
// 	h1 := discovery.NewMessageHandler(nil, nil)
// 	h2 := discovery.NewMessageHandler(h1, nil)
// 	h3 := discovery.NewMessageHandler(h2, consumer)

// 	//init with start handler
// 	pm := peer.NewManager(h3)

// 	//mock connection
// 	conn := network.NewLocalConn()

// 	//add seed peer connection
// 	pm.AddPeer(peer.NodeID("seed1"), conn)

// 	time.Sleep(3 * time.Second)

// 	//send packet to connection
// 	ping := discovery.NewPing(util.TimeNow(), 'c', 'd')

// 	payload := message.ToPayload(discovery.PingMessageType, ping)

// 	n, err := peer.Send(conn, payload, packet.COMPRESSED)

// 	log.Println(conn.Buf.Bytes(), n, err)

// 	//consume processed message
// 	msg := <-consumer

// 	log.Println(msg)
// }

// func test4() {
// 	// An artificial input source.
// 	const input = "Now is the winter of our discontent,\nMade glorious summer by this sun of York.\n"
// 	scanner := bufio.NewScanner(strings.NewReader(input))
// 	// Set the split function for the scanning operation.
// 	scanner.Split(bufio.ScanWords)
// 	// Count the words.
// 	count := 0
// 	for scanner.Scan() {
// 		count++
// 	}
// 	if err := scanner.Err(); err != nil {
// 		fmt.Fprintln(os.Stderr, "reading input:", err)
// 	}
// 	fmt.Printf("%d\n", count)
// }

// func test3() {
// 	var fleta fleta.Fleta
// 	mocknetwork.Fleta(&fleta)

// 	l, err := mocknet.Listen("tcp", "test")
// 	if err != nil {
// 		panic(err)
// 	}

// 	wg1 := sync.WaitGroup{}
// 	wg1.Add(2)
// 	wg2 := sync.WaitGroup{}
// 	wg2.Add(2)

// 	go func() {
// 		wg1.Done()
// 		conn, _ := mocknet.Dial("tcp", "test", "test1")
// 		fmt.Println("write1 start")
// 		data := []byte{10, 20}
// 		conn.Write(data)
// 		fmt.Println("write1 end")

// 		fmt.Println("read 2")
// 		time.Sleep(time.Second * 2)
// 		conn.Read(data)
// 		fmt.Println("data : ", data)

// 		wg2.Done()
// 	}()

// 	go func() {
// 		wg1.Done()
// 		conn, _ := l.Accept()
// 		fmt.Println("read 1")
// 		time.Sleep(time.Second * 2)
// 		data := make([]byte, 2)
// 		conn.Read(data)
// 		fmt.Println("data : ", data)

// 		fmt.Println("write2 start")
// 		data[0], data[1] = data[1], data[0]
// 		conn.Write(data)
// 		fmt.Println("write2 end")

// 		wg2.Done()
// 	}()

// 	wg1.Wait()
// 	wg2.Wait()

// 	fmt.Println("go end")
// }

// func test2() {
// 	sRead, _ := io.Pipe()
// 	_, sWrite := io.Pipe()

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	go func() {
// 		fmt.Println("write start")
// 		sWrite.Write([]byte{10, 20})
// 		fmt.Println("write end")
// 		wg.Done()
// 	}()

// 	go func() {
// 		fmt.Println("read 1")
// 		time.Sleep(time.Second)
// 		fmt.Println("read 2")
// 		var data []byte
// 		sRead.Read(data)
// 		fmt.Println("data : ", data)
// 		wg.Done()
// 	}()

// 	wg.Wait()

// 	fmt.Println("go end")
// }

// func test1() {
// 	l, err := net.Listen("tcp", ":3000")
// 	if err != nil {
// 		panic(err)
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	go func() {
// 		conn, _ := net.Dial("tcp", "127.0.0.1:3000")
// 		fmt.Println("write1 start")
// 		data := []byte{10, 20}
// 		conn.Write(data)
// 		fmt.Println("write1 end")

// 		fmt.Println("read 2")
// 		time.Sleep(time.Second * 2)
// 		conn.Read(data)
// 		fmt.Println("data : ", data, " conn : ", conn.RemoteAddr().String())

// 		wg.Done()
// 	}()

// 	go func() {
// 		conn, _ := l.Accept()
// 		fmt.Println("read 1")
// 		time.Sleep(time.Second * 2)
// 		data := make([]byte, 2)
// 		conn.Read(data)
// 		fmt.Println("data : ", data)

// 		fmt.Println("write2 start")
// 		data[0], data[1] = data[1], data[0]
// 		conn.Write(data)
// 		fmt.Println("write2 end")

// 		wg.Done()
// 	}()

// 	wg.Wait()

// 	fmt.Println("go end")

// }
