import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Service } from "../../../service"

const httpOptions = {
  headers: new HttpHeaders({
    'Content-Type':  'application/x-www-form-urlencoded',
    'Authorization': localStorage.getItem('access_token')
  })
};

@Component({
  selector: 'app-UserList',
  templateUrl: './UserList.component.html',
  styleUrls: ['./UserList.component.css']
})
export class UserListComponent implements OnInit {
  editor = false;
  initLoading = true;
  list: any[] = [];
  editorInfo={username:"", email:"", phonenumber:"", password:"", isAdmin:"false"}

  constructor(private http: HttpClient, private router: Router, private service:Service) {}
  userNameValidator() {
    var telStr = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
    if (!(telStr.test(this.editorInfo.username))) {
      return false
    }else{
      return true
	  }
  };

  checkEmail() {
    var telStr = /^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$/;
    if (!(telStr.test(this.editorInfo.email))) {
      return false
    }else{
      return true
    }
  };

  checkPN() {
    var telStr = /^[1](([3][0-9])|([4][5-9])|([5][0-3,5-9])|([6][5,6])|([7][0-8])|([8][0-9])|([9][1,8,9]))[0-9]{8}$/;
    if (!(telStr.test(this.editorInfo.phonenumber))) {
      return false
    }else{
       return true
    }
  };

  update(){
    if(this.checkEmail()&&this.userNameValidator()&&this.checkPN()){
      if(!this.service.updateService(this.editorInfo)){
        location.reload();
      }else{
        window.alert('输入的信息不正确')
      }
    }else{
      window.alert('输入的邮箱或手机号不正确')
    }

  }

  delete(item: any){
    this.service.deleteService(item)
  }

  logout(){
    this.service.logoutService()
  }

  onSameUrlNavigation: 'reload'
  ngOnInit(): void {
    if ( !this.service.ADMINaccess() ){
      this.router.navigateByUrl('dashboard');
    }
    this.http.get("http://localhost:3000/userlist",httpOptions).subscribe((data: any) => {
      this.list = data;
      this.initLoading = false;
    });
  }

  edit(item: any): void {
    this.editor=!this.editor;
    this.editorInfo.username=item.username;
    this.editorInfo.email=item.email;
    this.editorInfo.phonenumber=item.phonenumber;
  }
  cancel(){
    this.editor=!this.editor;
  }
}
