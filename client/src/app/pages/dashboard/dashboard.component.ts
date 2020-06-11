import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Service } from "../../service"

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent implements OnInit {

  constructor(private router: Router, private service:Service) { }

  ngOnInit() {
  }
  
  logout(){
    this.service.logoutService()
  }
}
