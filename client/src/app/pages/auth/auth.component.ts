import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

import { HttpClient, HttpParams } from '@angular/common/http';
import { Service } from "../../service"


@Component({
  selector: 'app-auth',
  templateUrl: './auth.component.html',
  styleUrls: ['./auth.component.css']
})

@Injectable()
export class AuthComponent implements OnInit {
  validateForm: FormGroup;

  submitForm(): void {
    for (const i in this.validateForm.controls) {
      this.validateForm.controls[i].markAsDirty();
      this.validateForm.controls[i].updateValueAndValidity();
    }
  }

  resolved(captchaResponse: string) {
    this.service.reCAPTCHA(captchaResponse)
  }

  constructor(private fb: FormBuilder, private router: Router, private service:Service, private http: HttpClient) { }

  passwordVisible = false;

  auth(){
    const data=new HttpParams()
    .set('username', this.validateForm.value.userName)
    .set('password', this.validateForm.value.passWord)
    this.service.authService(data)
  }

  ngOnInit(): void {
    this.validateForm = this.fb.group({
      userName: [null, [Validators.required]],
      passWord: [null, [Validators.required]],
      remember: [true]
    });
    if (this.service.logged()){
      this.router.navigateByUrl('/userMgr/dashboard');
    }
  }
}
