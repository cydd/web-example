import { Component, OnInit, TemplateRef } from '@angular/core';
import { Service } from "../../service"
import { HttpClient } from '@angular/common/http';


@Component({
  selector: 'app-message',
  templateUrl: './message.component.html',
  styleUrls: ['./message.component.css']
})
export class MessageComponent implements OnInit {

  messages = [];

  constructor(public service:Service, private http: HttpClient) {
  }

  chatTitle:string="选择一个用户进行私信"
  chatuser:string
  sendMessage(event) {
    this.service.sendmessage(event, this.messages, this.chatuser)
  }

  chat(item: any){
    this.chatTitle="私信给："+item.username
    this.chatuser=item.username
    this.messages.length = 0
    this.service.gethist(this.messages,this.chatuser)
    this.getmessage(this.messages,this.chatuser)
  }

  getmessage(messages:any[],chatuser:string){
    this.service.WebSocket.subscribe(res => {
      var t = 0;
      if(res['file']==true){
        t = 1;
      }
      messages.push({
        text: res['content'],
        date: res['date'],
        files: [ { url: res['file.src'], type: 'image/png' } ],
        type: t ? 'text' : 'file',
        reply: false,
        user: {
          name: res['sender'],
          avatar: 'https://i.gifer.com/no.gif',
        },
      })
    });
  }

  list: any[] = [];
  initLoading = true;
  ngOnInit() {
    this.http.get("http://localhost:3000/chatlist").subscribe((data: any) => {
      this.list = data;
      this.initLoading = false;
    });
    this.service.ws()
  }

  logout(){
    this.service.logoutService()
  }

}
