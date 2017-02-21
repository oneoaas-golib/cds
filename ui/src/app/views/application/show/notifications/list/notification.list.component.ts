import {Component, Input, Output, EventEmitter, ViewChild, AfterViewInit, OnDestroy} from '@angular/core';
import {Table} from '../../../../../shared/table/table';
import {Project} from '../../../../../model/project.model';
import {Application} from '../../../../../model/application.model';
import {Notification} from '../../../../../model/notification.model';
import {NotificationEvent} from '../notification.event';
import {ApplicationNotificationFormModalComponent} from '../form/notification.form.component';
import {Subscription} from 'rxjs/Rx';

declare var _: any;

@Component({
    selector: 'app-notification-list',
    templateUrl: './notification.list.html',
    styleUrls: ['./notification.list.scss']
})
export class ApplicationNotificationListComponent extends Table implements AfterViewInit, OnDestroy {

    @Input() notifications: Array<Notification>;
    @Input() edit = false;
    @Input() project: Project;
    @Input() application: Application;
    @Input() loading = false;

    @Output() event = new EventEmitter<NotificationEvent>();

    @ViewChild('notifForm')
    editNotifModal: ApplicationNotificationFormModalComponent;
    modalSubscription: Subscription;

    selectedNotification: Notification;

    constructor() {
        super();
    }

    ngOnDestroy(): void {
        if (this.modalSubscription) {
            this.modalSubscription.unsubscribe();
        }
    }

    ngAfterViewInit(): void {
        this.modalSubscription = this.editNotifModal.modal.onModalHide.subscribe( b => {
            if (b) {
                delete this.selectedNotification;
            }
        });
    }

    getData(): any[] {
        return this.notifications;
    }

    sendEvent(type: string, n: Notification) {
        this.loading = true;
        let notifs = new Array<Notification>();
        notifs.push(n);
        this.send(new NotificationEvent(type, notifs));

    }

    send(ne: NotificationEvent) {
        this.event.emit(ne);
    }

    openModal(n: Notification, key: string) {
        if (n) {
            this.selectedNotification = new Notification();
            this.selectedNotification.pipeline = _.cloneDeep(n.pipeline);
            this.selectedNotification.environment = _.cloneDeep(n.environment);
            this.selectedNotification.application_pipeline_id = n.application_pipeline_id;
            this.selectedNotification.notifications[key] = _.cloneDeep(n.notifications[key]);
        }
        this.editNotifModal.show({autofocus: false, closable: false, observeChanges: false});
    }

    close(): void {
        this.editNotifModal.close();
    }
}
