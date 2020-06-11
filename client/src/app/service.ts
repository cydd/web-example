import { HttpClient, HttpParams, HttpHeaders} from '@angular/common/http';
import { Router } from '@angular/router';
import { Injectable } from '@angular/core';
import { JwtHelperService } from '@auth0/angular-jwt';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';

const httpOptions = {
  headers: new HttpHeaders({
    'Content-Type':  'application/x-www-form-urlencoded',
    'Authorization': localStorage.getItem('access_token')
  })
};
const authOptions = {
  headers: new HttpHeaders({
    'Content-Type':  'application/x-www-form-urlencoded'
  })
};

@Injectable({
  providedIn: 'root'
})
export class Service {
  reCAPTCHAStatus:boolean=false

  constructor(private http: HttpClient, private router: Router,
    public jwtHelper: JwtHelperService) {}

  logged():boolean{
    if (localStorage.getItem('access_token')!=null) {
      return true
    }else{
      return false
    }
  }

  updateService(editorInfo:any):any{
    var Res:string="";
    const data=new HttpParams()
    .set('username', editorInfo.username)
    .set('password', editorInfo.password)
    .set('phonenumber', editorInfo.phonenumber)
    .set('email', editorInfo.email)

    this.http.post<any>("http://localhost:3000/updateUser", data, httpOptions)
      .subscribe((data: string) => {
      Res=data['Status']
      if (Res == "Succeeded"){
        return true
      }else{
        return false
      }
    })
  }

  deleteService(item: any){
    var Res:string="";
    const data=new HttpParams()
    .set('username', item.username)

    this.http.post<any>("http://localhost:3000/deleteUser", data, httpOptions)
      .subscribe((data: string) => {
      Res=data['Status']
      if (Res == "Succeeded"){
        window.alert('删除成功')
        location.reload();
      }
    })
  }

  logoutService(){
    localStorage.removeItem('access_token');
    this.router.navigateByUrl('/login');
    this.reCAPTCHAStatus==false
  }

  registerService(data:HttpParams){
    this.reCAPTCHAcheck()
    var Res:string="";
    this.http.post<any>("http://localhost:3000/addUser", data)
      .subscribe((data: string) => {
      Res=data['Status']
      if (Res == "Succeeded"){
        Res = data['token']
        localStorage.setItem('access_token',Res);
        this.router.navigateByUrl('/userMgr/dashboard');
      }
    })
  }

  authService(data:HttpParams){
    this.reCAPTCHAcheck()
    var Res:string="";
    this.http.post<any>("http://localhost:3000/auth", data, authOptions)
    .subscribe((data: string) => {
      Res = data['Status']
      if (Res == "Verified"){
        Res = data['token']
        localStorage.setItem('access_token',Res);
        this.router.navigateByUrl('/userMgr/dashboard')
        this.reCAPTCHAStatus==false
     }
    })
  }
  
  addService(data:HttpParams){
    var Res:string="";
    this.http.post<any>("http://localhost:3000/addUser", data, httpOptions)
      .subscribe((data: string) => {
      Res=data['Status']
      if (Res == "Succeeded"){
        this.router.navigateByUrl('/userMgr/userlist');
      }
    })
  }

  reCAPTCHA(captchaResponse: string){
    this.reCAPTCHAStatus=true;
    /*const data=new HttpParams()
    .set('captchaResponse', captchaResponse)

    this.http.post<any>("http://localhost:3000/reCAPTCHA", data)
      .subscribe((data: string) => {
      this.reCAPTCHAStatus=true;
    })*/
  }

  reCAPTCHAcheck(){
    if(this.reCAPTCHAStatus==false){
      window.alert('请先验证reCAPTCHA')
      location.reload();
    }
  }
  ADMINaccess():string{
    var temp=this.jwtHelper.decodeToken(localStorage.getItem('access_token'))
    return temp["isAdmin"]
  }


  WebSocket: WebSocketSubject<any>

  ws(){
    var temp:string="wss://cst.azeee.com/wss?username="+this.getusername()
    this.WebSocket= webSocket(temp)
  }

  getusername(){
    var username=this.jwtHelper.decodeToken(localStorage.getItem('access_token'))['username']
    return username
  }

  gethist(messages:any[],chatuser:string){
    var re:Boolean=true
    var username:string
    this.http.get<any>("http://localhost:3000/chathst",{params:{"user1":chatuser,"user2":this.getusername()}}).subscribe((data: any) => {
      for(var i=0;i<data.length;i++){
        if(chatuser==data[i]['sender']){
          re=false
          username=data[i]['sender']
        }else{
          re=true
          username=data[i]['sender']
        }
        var t = 0;
        if(data[i]['file']=='true'){
          t = 1;
        }
        messages.push({
          text: data[i]['content'],
          date: data[i]['date'],
          files:[{url:'https://cst.azeee.com'+data[i]['filesrc'],type:'image/png'}],
          type: t ? 'file' : 'text',
          reply: re,
          user: {
            name: username,
            avatar: 'https://i.gifer.com/no.gif',
          },
        });
      }
    });
  }

  sendmessage(event,messages:any[],chatuser:string){
    const files = !event.files ? [] : event.files.map((file) => {
      return {
        url: file.src,
        type: file.type,
        icon: 'file-text-outline',
      };
    })
    var username=this.getusername()
    var date=new Date()

    messages.push({
      text: event.message,
      date: date,
      files: files,
      type: files.length ? 'file' : 'text',
      reply: true,
      user: {
        name: username,
        avatar: 'https://i.gifer.com/no.gif',
      },
    });

    if (files.length) {
      const formData = new FormData();
      var content=event.message+""
      formData.append('image', event.files[0]);
      this.http.post<any>("http://localhost:3000/upload", formData).subscribe((data: string) => {
        var test:string='{"sender":"'+this.getusername()+'","receiver":"'+chatuser+
        '","content":"'+content+'","date":"'+date+'","file":"true","filesrc":"'+data['url']+'"}'
    
        var jsonStr = JSON.parse(test)
        this.WebSocket.next(jsonStr);
      });
    }else{
      var test:string='{"sender":"'+this.getusername()+'","receiver":"'+chatuser+
      '","content":"'+event.message+'","date":"'+date+'","file":"false","filesrc":"nofile"}'
  
      var jsonStr = JSON.parse(test)
      this.WebSocket.next(jsonStr);
    }
  }
}