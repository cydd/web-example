import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { AuthComponent } from './pages/auth/auth.component';
import { UserListComponent } from './pages/userMgr/userlist/userlist.component';
import { RegisterComponent } from './pages/register/register.component';
import { AdduserComponent } from './pages/userMgr/adduser/adduser.component';
import { adminDashboardComponent } from './pages/userMgr/dashboard/dashboard.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';
import { MessageComponent } from './pages/message/message.component'
import { Guard } from './guard'

@NgModule({
  imports: [
    RouterModule.forRoot([
    { path: '', redirectTo: '/login', pathMatch: 'full' },
    { path: 'login', component: AuthComponent },
    { path: 'dashboard', component: DashboardComponent, canActivate: [Guard] },
    { path: 'message', component: MessageComponent, canActivate: [Guard] },
    { path: 'userMgr/userlist', component: UserListComponent, canActivate: [Guard] },
    { path: 'register', component: RegisterComponent },
    { path: 'userMgr/adduser', component: AdduserComponent, canActivate: [Guard] },
    { path: 'userMgr/dashboard', component: adminDashboardComponent, canActivate: [Guard] },
  ])
],

  exports: [RouterModule]
})
export class AppRoutingModule { }
