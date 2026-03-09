export interface User {
  id: string;
  email: string;
  displayName: string;
  createdAt: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Meeting {
  id: string;
  title: string;
  description: string;
  organizerId: string;
  status: 'pending' | 'confirmed' | 'cancelled';
  isPublic: boolean;
  tags?: string[];
  confirmedSlotId?: string;
  createdAt: string;
  organizer?: User;
  timeSlots?: TimeSlot[];
  participants?: Participant[];
}

export interface TimeSlot {
  id: string;
  meetingId: string;
  startTime: string;
  endTime: string;
  voteCount: number;
  voters?: User[];
}

export interface Participant {
  id: string;
  meetingId: string;
  userId: string;
  rsvpStatus: 'invited' | 'accepted' | 'declined';
  user?: User;
}

export interface CreateMeetingInput {
  title: string;
  description: string;
  isPublic: boolean;
  tags: string[];
  timeSlots: { startTime: string; endTime: string }[];
  participantEmails: string[];
  participantIds: string[];
}

export interface VoteInput {
  timeSlotIds: string[];
}
