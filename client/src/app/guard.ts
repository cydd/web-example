import { Injectable } from '@angular/core';
import { Router, CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { JwtHelperService } from '@auth0/angular-jwt';

@Injectable()
export class Guard implements CanActivate {
  constructor(private router: Router, public jwtHelper: JwtHelperService) { }

  canActivate(next: ActivatedRouteSnapshot, state: RouterStateSnapshot,) {
    if (localStorage.getItem('access_token')) {
      if (this.jwtHelper.isTokenExpired(localStorage.getItem('access_token'))){
        localStorage.removeItem('access_token');
        return false;
      }
      return true;
    }

    this.router.navigate(['login']);
    return false;
  }
}