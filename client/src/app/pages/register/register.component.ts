import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, ValidationErrors, Validators } from '@angular/forms';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, Observer, range } from 'rxjs';

import { HttpClient, HttpParams, HttpHeaders } from '@angular/common/http';
import { Service } from "../../service"


const httpOptions = {
  headers: new HttpHeaders({
    'Content-Type':  'application/json',
    'Authorization': 'AngularPOST'
  })
};

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})

@Injectable()
export class RegisterComponent implements OnInit {

  validateForm: FormGroup;

  submitForm(): void {
    for (const i in this.validateForm.controls) {
      this.validateForm.controls[i].markAsDirty();
      this.validateForm.controls[i].updateValueAndValidity();
    }
  }

  userNameValidator = (control: FormControl) =>
  new Observable((observer: Observer<ValidationErrors | null>) => {
    var telStr = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
    if (!(telStr.test(this.validateForm.controls.UserName.value))) {
      observer.next({ error: true, checkC: true });
      observer.complete();
    }else{
      observer.next(null);
      observer.complete();
	}
  });

  checkEmail = (control: FormControl) =>
  new Observable((observer: Observer<ValidationErrors | null>) => {
    var telStr = /^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$/;
    if (!(telStr.test(this.validateForm.controls.email.value))) {
      observer.next({ error: true, email: true });
      observer.complete();
    }else{
      observer.next(null);
      observer.complete();
	}
  });

  checkPN = (control: FormControl) =>
  new Observable((observer: Observer<ValidationErrors | null>) => {
    var telStr = /^[1](([3][0-9])|([4][5-9])|([5][0-3,5-9])|([6][5,6])|([7][0-8])|([8][0-9])|([9][1,8,9]))[0-9]{8}$/;
    if (!(telStr.test(this.validateForm.controls.phoneNumber.value))) {
      observer.next({ error: true, phoneNumber: true });
      observer.complete();
    }else{
      observer.next(null);
      observer.complete();
	}
  });

  checkexisted():any{
    var Res :string ="";
    this.http.get("http://localhost:3000/checkinfo", 
    {params:new HttpParams()
    .set('username', this.validateForm.controls.UserName.value)
    .set('email', this.validateForm.controls.email.value)
    .set('phonenumber', this.validateForm.controls.phoneNumber.value)})
    .subscribe((data: string) => {
      Res = data['Status'] 
    setTimeout(() => {
        if (Res=="existed") {
          return true;
        } else {
          return false;
        }
    }, 1000)})
  }

  validateConfirmPassword(): void {
    setTimeout(() => this.validateForm.controls.confirm.updateValueAndValidity());
  }

  confirmValidator = (control: FormControl): { [s: string]: boolean } => {
    if (!control.value) {
      return { error: true, required: true };
    } else if (control.value !== this.validateForm.controls.password.value) {
      return { confirm: true, error: true };
    }
    return {};
  };

  resolved(captchaResponse: string) {
    this.service.reCAPTCHA(captchaResponse)
  }

  constructor(private fb: FormBuilder, private router: Router, private http: HttpClient, private service:Service) {}

  register(){
    if(!this.checkexisted()){
      const data=new HttpParams()
      .set('username', this.validateForm.value.UserName)
      .set('password', this.validateForm.value.Password)
      .set('phonenumber', this.validateForm.value.phoneNumber)
      .set('email', this.validateForm.value.email)
      .set('isAdmin', 'true')
      this.service.registerService(data)
    }else{
      window.alert('用户名或邮箱或手机号已存在')
    }
  }

  ngOnInit(): void {
    this.validateForm = this.fb.group({
      email: [null, [Validators.required], [this.checkEmail]],
      Password: [null, [Validators.required]],
      checkPassword: [null, [Validators.required]],
      phoneNumberPrefix: ['+86'],
      phoneNumber: [null, [Validators.required], [this.checkPN]], 
      UserName: [null, [Validators.required], [this.userNameValidator]],
      confirm: ['', [this.confirmValidator]],
      agree: [false, [Validators.required]]
    });
    if (this.service.logged()){
      this.router.navigateByUrl('/userMgr/dashboard');
    }
  }
}